package handlers

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
)

// LangHandlers Function creates language routes like /ru/someroute, /en/someroute etc.
func LangHandlers(hr *chi.Mux, route string, siteConfig *site.Config, handler http.Handler) {
	languages := internal.GetLanguages()
	for _, l := range languages {
		langId := l.Id
		var preparedRoute string
		if strings.Contains(route, "{lang}") {
			preparedRoute = strings.ReplaceAll(route, "{lang}", langId)
		} else {
			preparedRoute = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, "{lang}", langId)
			preparedRoute = strings.ReplaceAll(preparedRoute, "{route}", route)
		}
		if len(preparedRoute) > 1 {
			preparedRoute = strings.TrimSuffix(preparedRoute, "/")
		}
		hr.Handle(preparedRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "lang", langId)
			r.AddCookie(&http.Cookie{
				Name:     internal.Config.General.LangCookie,
				Value:    langId,
				Expires:  time.Now().Add(time.Hour * 24 * 30),
				SameSite: http.SameSiteLaxMode,
			})
			handler.ServeHTTP(w, r.WithContext(ctx))
		}))
		if !strings.Contains(route, "{lang}") {
			// And route to detect lang
			if route == siteConfig.Routes.TopCategories {
				hr.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					langCookie, _ := r.Cookie(internal.Config.General.LangCookie)
					hostName := r.Context().Value("hostName").(string)
					langValue := "en"
					if langCookie != nil {
						langValue = langCookie.Value
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
									for _, t := range cat.Tags {
										if strings.Contains(s, t) {
											redirectUrl := strings.ReplaceAll(preparedRoute, "{route}", siteConfig.Routes.Category)
											redirectUrl = strings.ReplaceAll(redirectUrl, "{slug}", cat.Slug)
											redirectUrl = strings.ReplaceAll(redirectUrl, "{id}", strconv.FormatInt(int64(cat.Id), 10))
											if qs := r.URL.RawQuery; qs != "" {
												redirectUrl = redirectUrl + "?" + qs
											}
											http.Redirect(w, r, redirectUrl, 302)
											return
										}
									}
								}
							}
						}
					}
					lang := internal.DetectLanguage(langValue, r.Header.Get("Accept-Language"))
					var redirectUri string
					if lang == nil {
						redirectUri = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, "{lang}", "en")
					} else {
						redirectUri = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, "{lang}", lang.Id)
					}
					redirectUri = strings.ReplaceAll(redirectUri, "{route}", route)
					http.Redirect(w, r, redirectUri, 302)
					return
				}))
			} else {
				hr.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					langCookie, _ := r.Cookie(internal.Config.General.LangCookie)
					langValue := "en"
					if langCookie != nil {
						langValue = langCookie.Value
					}
					lang := internal.DetectLanguage(langValue, r.Header.Get("Accept-Language"))
					var redirectUri string
					if lang == nil {
						redirectUri = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, "{lang}", "en")
					} else {
						redirectUri = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, "{lang}", lang.Id)
					}
					var uri = r.URL.Path
					if r.URL.RawQuery != "" {
						uri += "?" + r.URL.RawQuery
					}
					redirectUri = strings.ReplaceAll(redirectUri, "{route}", uri)
					http.Redirect(w, r, redirectUri, 302)
					return
				}))
			}
		}
	}

}
