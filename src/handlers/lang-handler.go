package handlers

import (
	"context"
	"github.com/samber/lo"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
)

// LangHandlers Function creates language routes like /ru/someroute, /en/someroute etc.
func LangHandlers(hr *chi.Mux, route string, siteConfig *types.Config, handler http.Handler) {
	languages := internal.GetLanguages()
	var langTemplate = siteConfig.Routes.LanguageTemplate
	for k, v := range siteConfig.Routes.Custom {
		if v == route {
			if langTemplateCustom, ok := siteConfig.Routes.Custom[k+"_multilang"]; ok {
				langTemplate = langTemplateCustom
			}
			break
		}
	}
	for _, l := range languages {
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
					return
				}
				if len(matches) > 0 {
					w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
					w.Header().Add("Pragma", "no-cache")
					http.Redirect(w, r, route, http.StatusMovedPermanently)
					return
				}
			}
			handler.ServeHTTP(w, r.WithContext(ctx))
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
									redirectUrl = strings.ReplaceAll(redirectUrl, "{id}", strconv.FormatInt(int64(cat.Id), 10))
									if qs := r.URL.RawQuery; qs != "" {
										redirectUrl = redirectUrl + "?" + qs
									}
									w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
									w.Header().Add("Pragma", "no-cache")
									http.Redirect(w, r, redirectUrl, 302)
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
				handler.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			redirectUri = strings.ReplaceAll(redirectUri, "{route}", route)
			if qs := r.URL.RawQuery; qs != "" {
				redirectUri = redirectUri + "?" + qs
			}
			w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Add("Pragma", "no-cache")
			http.Redirect(w, r, redirectUri, 302)
			return
		}))
	} else {
		hr.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			langCookie, _ := r.Cookie(internal.Config.General.LangCookie)
			langValue := ""
			if langCookie != nil {
				langValue = langCookie.Value
			}
			lang := internal.DetectLanguage(langValue, siteConfig.General.DefaultLanguage, r.Header.Get("Accept-Language"))
			if siteConfig.General.NoRedirectDefaultLanguage && lang.Id == siteConfig.General.DefaultLanguage {
				ctx := context.WithValue(r.Context(), "lang", lang.Id)
				ctx = context.WithValue(ctx, "isXDefault", true)
				handler.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			var redirectUri string
			redirectUri = strings.ReplaceAll(langTemplate, "{lang}", lang.Id)
			var uri = r.URL.Path
			redirectUri = strings.ReplaceAll(redirectUri, "{route}", uri)
			if r.URL.RawQuery != "" {
				redirectUri += "?" + r.URL.RawQuery
			}
			w.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Add("Pragma", "no-cache")
			http.Redirect(w, r, redirectUri, 302)
			return
		}))
	}
}
