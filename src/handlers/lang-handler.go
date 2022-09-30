package handlers

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
)

// LangHandlers Function creates language routes like /ru/someroute, /en/someroute etc.
func LangHandlers(app *fiber.App, route string, siteConfig *site.Config, handlers ...fiber.Handler) {
	languages := internal.GetLanguages()
	for _, l := range languages {
		langId := l.Id
		var r string
		if strings.Contains(route, ":lang") {
			r = strings.ReplaceAll(route, ":lang", langId)
		} else {
			r = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, ":lang", langId)
			r = strings.ReplaceAll(r, ":route", route)
		}
		h := append([]fiber.Handler{func(c *fiber.Ctx) error {
			c.Locals("lang", langId)
			c.Cookie(&fiber.Cookie{
				Name:     internal.Config.General.LangCookie,
				Value:    langId,
				Expires:  time.Now().Add(time.Hour * 24 * 30),
				SameSite: "lax",
			})
			return c.Next()
		}}, handlers...)
		app.All(r, h...)
	}
	if !strings.Contains(route, ":lang") {
		// And route to detect lang
		if route == siteConfig.Routes.TopCategories {
			app.All(route, func(c *fiber.Ctx) error {
				langCookie := c.Cookies(internal.Config.General.LangCookie)
				hostName := c.Locals("hostName").(string)
				lang := internal.DetectLanguage(langCookie, c.Get("Accept-Language"))
				var r string
				if lang == nil {
					r = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, ":lang", "en")
				} else {
					r = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, ":lang", lang.Id)
				}
				if ref := c.Get("Referer"); ref != "" {
					if u, err := url.Parse(ref); err == nil &&
						strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") != hostName &&
						!botDetector.IsBot(c.Get("User-Agent")) {
						var s = strings.ToLower(u.Path + " " + u.RawQuery)
						if categories, err := db.GetCachedTopCategories(hostName); err == nil {
							for _, cat := range categories.Items {
								for _, t := range cat.Tags {
									if strings.Contains(s, t) {
										redirectUrl := strings.ReplaceAll(r, ":route", siteConfig.Routes.Category)
										redirectUrl = strings.ReplaceAll(redirectUrl, ":slug", cat.Slug)
										redirectUrl = strings.ReplaceAll(redirectUrl, ":id", strconv.FormatInt(int64(cat.Id), 10))
										if qs := string(c.Request().URI().QueryString()); qs != "" {
											redirectUrl = redirectUrl + "?" + qs
										}
										return c.Redirect(redirectUrl)
									}
								}
							}
						}
					}
				}
				r = strings.ReplaceAll(r, ":route", route)
				return c.Redirect(r)
			})
		} else {
			app.All(route, func(c *fiber.Ctx) error {
				langCookie := c.Cookies(internal.Config.General.LangCookie)
				lang := internal.DetectLanguage(langCookie, c.Get("Accept-Language"))
				var r string
				if lang == nil {
					r = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, ":lang", "en")
				} else {
					r = strings.ReplaceAll(siteConfig.Routes.LanguageTemplate, ":lang", lang.Id)
				}
				r = strings.ReplaceAll(r, ":route", route)
				return c.Redirect(r)
			})
		}
	}
}
