package handlers

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/types"

	"github.com/go-chi/chi/v5"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
)

var urlRegex = regexp.MustCompile(`^https?://([^/]+)`)

// LangHandlers Function creates language routes like /ru/someroute, /en/someroute etc.
func LangHandlers(hr *chi.Mux, route string, siteConfig *types.Config, handler http.Handler) {
	allLanguages := internal.GetLanguages(nil)
	languages := internal.GetLanguages(siteConfig)
	var langMap = make(map[string]struct{})
	for _, l := range languages {
		langMap[l.Id] = struct{}{}
	}
	var langTemplate = siteConfig.Routes.LanguageTemplate
	for k, v := range siteConfig.Routes.Custom {
		if v == route {
			if langTemplateCustom, ok := siteConfig.Routes.Custom[k+"_multilang"]; ok {
				langTemplate = langTemplateCustom
			}
			break
		}
	}
	for _, l := range allLanguages {
		langId := l.Id
		var preparedRoute string
		if strings.Contains(route, "{lang}") {
			preparedRoute = strings.ReplaceAll(route, "{lang}", langId)
		} else {
			preparedRoute = strings.ReplaceAll(langTemplate, "{lang}", langId)
			preparedRoute = strings.ReplaceAll(preparedRoute, "{route}", route)
		}
		if len(preparedRoute) > 1 {
			preparedRoute = strings.TrimSuffix(preparedRoute, "/")
		}
		hr.Handle(preparedRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if d, ok := siteConfig.LanguageDomains[langId]; ok {
				matches := urlRegex.FindStringSubmatch(d)
				var scheme = "https://"
				var uri = "/"
				if len(matches) > 1 {
					u, _ := url.Parse(d)
					scheme = u.Scheme + "://"
					uri = u.Path
					d = u.Hostname()
				}
				if d != r.Context().Value("hostName") {
					uriToRedirect := path.Join(uri, r.URL.Path)
					if r.URL.RawQuery != "" {
						uriToRedirect += "?" + r.URL.RawQuery
					}
					http.Redirect(w, r, scheme+d+uriToRedirect, http.StatusMovedPermanently)
					return
				}
			} else if siteConfig.General.CanonicalUrl != "" {
				// detect canonicalDomain
				canonicalParsed, _ := url.Parse(siteConfig.General.CanonicalUrl)
				canonicalDomain := canonicalParsed.Hostname()
				canonicalDomain = strings.TrimPrefix(canonicalDomain, "www.")
				if canonicalDomain != r.Context().Value("hostName") {
					uriToRedirect := r.URL.Path
					if r.URL.RawQuery != "" {
						uriToRedirect += "?" + r.URL.RawQuery
					}
					http.Redirect(w, r, "https://"+canonicalParsed.Host+uriToRedirect, http.StatusMovedPermanently)
					return
				}
			}
			// Проверяем, входит ли язык в список поддерживаемых
			_, isLanguageSupported := langMap[langId]
			if !isLanguageSupported {
				// Редирект на страницу с дефолтным языком
				defaultLang := siteConfig.General.DefaultLanguage
				var redirectUri string

				// Формируем путь с дефолтным языком
				if strings.Contains(route, "{lang}") {
					redirectUri = strings.ReplaceAll(route, "{lang}", defaultLang)
				} else {
					redirectUri = strings.ReplaceAll(langTemplate, "{lang}", defaultLang)
					redirectUri = strings.ReplaceAll(redirectUri, "{route}", route)
				}

				if len(redirectUri) > 1 {
					redirectUri = strings.TrimSuffix(redirectUri, "/")
				}
				// нам надо параметры тоже переписать в урле, типа {slug}, {id} и т.д.
				ctx := chi.RouteContext(r.Context())
				for i, key := range ctx.URLParams.Keys {
					value := ctx.URLParams.Values[i]
					redirectUri = strings.ReplaceAll(redirectUri, "{"+key+"}", value)
				}
				// Сохраняем параметры запроса
				if r.URL.RawQuery != "" {
					redirectUri += "?" + r.URL.RawQuery
				}

				http.Redirect(w, r, redirectUri, http.StatusMovedPermanently)
				if internal.Config.General.EnableAccessLog {
					log.Println(r.Context().Value("hostName").(string), 301, "Unsupported language redirect to", redirectUri)
				}
				return
			}
			ctx := context.WithValue(r.Context(), "lang", langId)
			ctx = context.WithValue(ctx, "isXDefault", false)
			cookie := &http.Cookie{
				Name:     internal.Config.General.LangCookie,
				Value:    langId,
				Expires:  time.Now().Add(time.Hour * 24 * 30),
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			}
			r.AddCookie(cookie)
			http.SetCookie(w, cookie)
			config := r.Context().Value("config").(*types.Config)
			if config.General.NoRedirectDefaultLanguage && langId == config.General.DefaultLanguage && !strings.Contains(route, "{lang}") {
				// for default language, we need to redirect to x-default uri
				escapedTemplate := regexp.QuoteMeta(langTemplate)
				regexPattern := strings.ReplaceAll(escapedTemplate, "\\{route\\}", "(.*)")
				regexPattern = strings.ReplaceAll(regexPattern, "\\{lang\\}", config.General.DefaultLanguage)
				re := regexp.MustCompile("^" + regexPattern + "$")
				matches := re.FindStringSubmatch(r.URL.Path)
				if len(matches) > 1 {
					// redirect to x-default
					w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
					w.Header().Add("Pragma", "no-cache")
					http.Redirect(w, r, matches[1], http.StatusMovedPermanently)
					if internal.Config.General.EnableAccessLog {
						log.Println(r.Context().Value("hostName").(string), 301, matches[1])
					}
					return
				}
				if len(matches) > 0 {
					w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
					w.Header().Add("Pragma", "no-cache")
					http.Redirect(w, r, route, http.StatusMovedPermanently)
					if internal.Config.General.EnableAccessLog {
						log.Println(r.Context().Value("hostName").(string), 301, route)
					}
					return
				}
			}
			middlewares.BadBotMiddleware(handler).ServeHTTP(w, r.WithContext(ctx))
		}))
	}
	if strings.Contains(route, "{lang}") {
		// if route has {lang} - it will not have x-default and language detect version
		return
	}
	// And route to detect lang
	if route == siteConfig.Routes.TopCategories {
		hr.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			langCookie, _ := r.Cookie(internal.Config.General.LangCookie)
			hostName := r.Context().Value("hostName").(string)
			langValue := ""
			if langCookie != nil {
				langValue = langCookie.Value
			}
			lang := internal.DetectLanguage(langValue, siteConfig.General.DefaultLanguage, r.Header.Get("Accept-Language"))
			// Проверяем, что обнаруженный язык поддерживается
			_, isLanguageSupported := langMap[lang.Id]
			if !isLanguageSupported {
				// Используем дефолтный язык вместо обнаруженного
				lang = internal.GetLanguage(siteConfig.General.DefaultLanguage)
			}
			var redirectUri string
			if lang.Id == siteConfig.General.DefaultLanguage && siteConfig.General.NoRedirectDefaultLanguage {
				redirectUri = "{route}"
			} else {
				redirectUri = strings.ReplaceAll(langTemplate, "{lang}", lang.Id)
			}
			
			ip := r.Context().Value("ip").(string)
			groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
			if ref := r.Header.Get("Referer"); ref != "" {
				if u, err := url.Parse(ref); err == nil &&
					strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") != hostName &&
					!botDetector.IsBot(r.Header.Get("User-Agent")) {
					var s = strings.ToLower(u.Path + " " + u.RawQuery)
					if categories, err := db.GetCachedTopCategories(hostName, groupId); err == nil {
						for _, cat := range categories.Items {
							tags := make([]string, 0, len(cat.Tags)+1)
							for _, t := range cat.Tags {
								tags = append(tags, strings.ToLower(t))
							}
							title := strings.ToLower(cat.Title)
							if !lo.Contains(tags, title) {
								tags = append(tags, title)
							}
							for _, t := range tags {
								if strings.Contains(s, t) {
									redirectUrl := strings.ReplaceAll(redirectUri, "{route}", siteConfig.Routes.Category)
									redirectUrl = strings.ReplaceAll(redirectUrl, "{slug}", cat.Slug)
									catId := int64(cat.Id)
									if siteConfig.Routes.IdXorKey > 0 {
										catId = catId ^ siteConfig.Routes.IdXorKey
									}
									redirectUrl = strings.ReplaceAll(redirectUrl, "{id}", strconv.FormatInt(catId, 10))
									if qs := r.URL.RawQuery; qs != "" {
										redirectUrl = redirectUrl + "?" + qs
									}
									w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
									w.Header().Add("Pragma", "no-cache")
									http.Redirect(w, r, redirectUrl, http.StatusFound)
									if internal.Config.General.EnableAccessLog {
										log.Println(hostName, 302, redirectUrl)
									}
									return
								}
							}
						}
					}
				}
			}
			if siteConfig.General.NoRedirectDefaultLanguage && lang.Id == siteConfig.General.DefaultLanguage {
				ctx := context.WithValue(r.Context(), "lang", lang.Id)
				ctx = context.WithValue(ctx, "isXDefault", true)
				middlewares.BadBotMiddleware(handler).ServeHTTP(w, r.WithContext(ctx))
				return
			}
			redirectUri = strings.ReplaceAll(redirectUri, "{route}", route)
			if qs := r.URL.RawQuery; qs != "" {
				redirectUri = redirectUri + "?" + qs
			}
			w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Add("Pragma", "no-cache")
			http.Redirect(w, r, redirectUri, http.StatusFound)
			if internal.Config.General.EnableAccessLog {
				log.Println("Redirected", hostName, 302, redirectUri)
			}
		}))
	} else {
		hr.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			langCookie, _ := r.Cookie(internal.Config.General.LangCookie)
			langValue := ""
			if langCookie != nil {
				langValue = langCookie.Value
			}
			lang := internal.DetectLanguage(langValue, siteConfig.General.DefaultLanguage, r.Header.Get("Accept-Language"))
			// Проверяем, что обнаруженный язык поддерживается
			_, isLanguageSupported := langMap[lang.Id]
			if !isLanguageSupported {
				// Используем дефолтный язык вместо обнаруженного
				lang = internal.GetLanguage(siteConfig.General.DefaultLanguage)
			}
			var redirectUri string
			redirectUri = strings.ReplaceAll(langTemplate, "{lang}", lang.Id)
			var uri = r.URL.Path
			redirectUri = strings.ReplaceAll(redirectUri, "{route}", uri)
			if siteConfig.General.NoRedirectDefaultLanguage && lang.Id == siteConfig.General.DefaultLanguage || redirectUri == uri {
				ctx := context.WithValue(r.Context(), "lang", lang.Id)
				ctx = context.WithValue(ctx, "isXDefault", true)
				middlewares.BadBotMiddleware(handler).ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if r.URL.RawQuery != "" {
				redirectUri += "?" + r.URL.RawQuery
			}
			w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Add("Pragma", "no-cache")
			http.Redirect(w, r, redirectUri, http.StatusFound)
			if internal.Config.General.EnableAccessLog {
				log.Println("Redirected to ", redirectUri)
			}
		}))
	}
}
