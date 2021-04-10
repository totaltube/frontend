package handlers

import (
	"github.com/gofiber/fiber/v2"
	"sersh.com/totaltube/frontend/internal"
	"strings"
	"time"
)

// Function creates language routes like /ru/someroute, /en/someroute etc.
func LangHandlers(app *fiber.App, route string, langTemplate string, handler fiber.Handler) {
	languages := internal.GetLanguages()
	for _, l := range languages {
		langId := l.Id
		var r string
		if strings.Contains(route, ":lang") {
			r = strings.ReplaceAll(route, ":lang", langId)
		} else {
			r = strings.ReplaceAll(langTemplate, ":lang", langId)
			r = strings.ReplaceAll(r, ":route", route)
		}
		app.All(r, func(c *fiber.Ctx) error {
			c.Locals("lang", langId)
			c.Cookie(&fiber.Cookie{
				Name: internal.Config.General.LangCookie,
				Value: langId,
				Expires: time.Now().Add(time.Hour*24*30),
				SameSite: "lax",
			})
			return c.Next()
		}, handler)
	}
	if !strings.Contains(route, ":lang") {
		// And route to detect lang
		app.All(route, func(c *fiber.Ctx) error {
			langCookie := c.Cookies(internal.Config.General.LangCookie)
			lang := internal.DetectLanguage(langCookie, c.Get("Accept-Language"))
			var r string
			if lang == nil {
				r = strings.ReplaceAll(langTemplate, ":lang", "en")
			} else {
				r = strings.ReplaceAll(langTemplate, ":lang", lang.Id)
			}
			r = strings.ReplaceAll(r, ":route", route)
			return c.Redirect(r)
		})
	}
}
