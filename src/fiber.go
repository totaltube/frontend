package main

import (
	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	recoverMiddleware "github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"os"
	"path/filepath"
	"sersh.com/totaltube/frontend/handlers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strings"
)

type host struct {
	fiber  *fiber.App
	config *site.Config
	path   string
}

func InitFiber() *fiber.App {
	fiberConfig := fiber.Config{
		ProxyHeader:           internal.Config.General.RealIpHeader,
		CaseSensitive:         true,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) (result error) {
			defer func() {
				if r, ok := recover().(error); r != nil && ok {
					log.Println(r)
					if c.Accepts("application/json") != "" {
						result = c.JSON(map[string]interface{}{
							"success": false,
							"value":   err.Error(),
						})
						return
					}
					result = c.Status(fiber.StatusInternalServerError).SendString(err.Error())
				}
			}()
			if err == types.ErrResponseSent  {
				// response already sent
				return nil
			}
			if c.Accepts("text/html") != "" {
				return handlers.Generate500(c, err.Error())
				//return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}
			if c.Accepts("application/json") != "" {
				return c.JSON(map[string]interface{}{
					"success": false,
					"value":   err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	}
	hosts := map[string]*host{}
	defaultSiteOk := false
	matches, err := filepath.Glob(filepath.Join(internal.Config.Frontend.SitesPath, "*"))
	if err != nil {
		log.Fatalln(err)
	}
	for _, m := range matches {
		configPath := filepath.Join(m, "config.toml")
		if _, err := os.Stat(configPath); err != nil {
			continue
		}
		h := host{
			config: site.NewConfig(),
			path:   m,
		}
		hostName := filepath.Base(m)
		if _, err := toml.DecodeFile(configPath, h.config); err != nil {
			log.Fatalln("error reading site config at", configPath, err)
		}
		jsPath := filepath.Join(m, "js")
		if _, err := os.Stat(jsPath); err == nil {
			site.WatchJS(jsPath, h.config) // Следить за директорией и пересоздавать js
		}
		scssPath := filepath.Join(m, "scss")
		if _, err := os.Stat(scssPath); err == nil {
			site.WatchScss(scssPath, h.config) // Следить за scss и пересоздавать css
		}
		h.fiber = fiber.New(fiberConfig)
		h.fiber.Use(recoverMiddleware.New())
		h.fiber.Use(func(c *fiber.Ctx) error {
			c.Locals("config", h.config)
			c.Locals("path", h.path)
			c.Locals("hostName", hostName)
			c.Locals("lang", "en")
			return c.Next()
		})
		h.fiber.Static("/", filepath.Join(m, "public"), fiber.Static{
			Compress:  false,
			ByteRange: true,
			Browse:    false,
			MaxAge:    3600,
		})
		if h.config.Routes.Autocomplete != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Autocomplete, h.config.Routes.LanguageTemplate, handlers.Autocomplete)
			} else {
				h.fiber.All(h.config.Routes.Autocomplete, handlers.Autocomplete)
			}
		}
		if h.config.Routes.Search != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Search, h.config.Routes.LanguageTemplate, handlers.Search)
			} else {
				h.fiber.All(h.config.Routes.Search, handlers.Search)
			}
		}
		if h.config.Routes.Category != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Category, h.config.Routes.LanguageTemplate, handlers.Category)
			} else {
				h.fiber.All(h.config.Routes.Category, handlers.Category)
			}
		}
		if h.config.Routes.TopCategories != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.TopCategories, h.config.Routes.LanguageTemplate, handlers.TopCategories)
			} else {
				h.fiber.All(h.config.Routes.TopCategories, handlers.TopCategories)
			}
		}
		if h.config.Routes.TopContent != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.TopContent, h.config.Routes.LanguageTemplate, handlers.TopContent)
			} else {
				h.fiber.All(h.config.Routes.TopContent, handlers.TopContent)
			}
		}
		if h.config.Routes.Model != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Model, h.config.Routes.LanguageTemplate, handlers.Model)
			} else {
				h.fiber.All(h.config.Routes.Model, handlers.Model)
			}
		}
		if h.config.Routes.Channel != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Channel, h.config.Routes.LanguageTemplate, handlers.Channel)
			} else {
				h.fiber.All(h.config.Routes.Channel, handlers.Channel)
			}
		}
		if h.config.Routes.ContentItem != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.ContentItem, h.config.Routes.LanguageTemplate, handlers.ContentItem)
			} else {
				h.fiber.All(h.config.Routes.ContentItem, handlers.ContentItem)
			}
		}
		if h.config.Routes.New != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.New, h.config.Routes.LanguageTemplate, handlers.New)
			} else {
				h.fiber.All(h.config.Routes.New, handlers.New)
			}
		}
		if h.config.Routes.Long != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Long, h.config.Routes.LanguageTemplate, handlers.Long)
			} else {
				h.fiber.All(h.config.Routes.Long, handlers.Long)
			}
		}
		if h.config.Routes.Popular != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Popular, h.config.Routes.LanguageTemplate, handlers.Popular)
			} else {
				h.fiber.All(h.config.Routes.Popular, handlers.Popular)
			}
		}
		if h.config.Routes.Models != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Models, h.config.Routes.LanguageTemplate, handlers.Models)
			} else {
				h.fiber.All(h.config.Routes.Models, handlers.Models)
			}
		}
		if h.config.Routes.Out != "" {
			if h.config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, h.config.Routes.Out, h.config.Routes.LanguageTemplate, handlers.Out)
			} else {
				h.fiber.All(h.config.Routes.Out, handlers.Out)
			}
		}
		if h.config.Routes.Custom != nil {
			for templateName, routePath := range h.config.Routes.Custom {
				if h.config.General.MultiLanguage {
					handlers.LangHandlers(h.fiber, routePath, h.config.Routes.LanguageTemplate, func(c *fiber.Ctx) error {
						c.Locals("custom_template_name", templateName)
						return handlers.Custom(c)
					})
				} else {
					h.fiber.All(routePath, func(c *fiber.Ctx) error {
						c.Locals("custom_template_name", templateName)
						return c.Next()
					}, handlers.Custom)
				}
			}
		}
		hosts[hostName] = &h
		if hostName == internal.Config.Frontend.DefaultSite {
			defaultSiteOk = true
		}
	}
	if !defaultSiteOk {
		log.Fatalln("Default site", internal.Config.Frontend.DefaultSite,
			"not present in", internal.Config.Frontend.SitesPath)
	}

	app := fiber.New(fiberConfig)
	app.Use(recoverMiddleware.New())
	app.Use(func(c *fiber.Ctx) error {
		hostName := strings.TrimPrefix(c.Hostname(), "www.")
		host := hosts[hostName]
		if host == nil {
			host = hosts[internal.Config.Frontend.DefaultSite]
		}
		host.fiber.Handler()(c.Context())
		return nil
	})
	return app
}
