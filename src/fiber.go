package main

import (
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
	fiber      *fiber.App
	configPath string
	path       string
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
					if c.Accepts("application/json") != "" && c.Accepts("text/html") == "" {
						result = c.JSON(map[string]interface{}{
							"success": false,
							"value":   err.Error(),
						})
						return
					}
					result = c.Status(fiber.StatusInternalServerError).SendString(err.Error())
				}
			}()
			if err == types.ErrResponseSent {
				// response already sent
				return nil
			}
			log.Println("error on page", string(c.Request().RequestURI()), ":", err)
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
		config := site.GetConfigAndWatch(configPath)
		h := host{
			configPath: configPath,
			path:       m,
		}
		hostName := filepath.Base(m)
		jsPath := filepath.Join(m, "js")
		if _, err := os.Stat(jsPath); err == nil {
			site.WatchJS(jsPath, configPath) // Следить за директорией и пересоздавать js
		}
		scssPath := filepath.Join(m, "scss")
		if _, err := os.Stat(scssPath); err == nil {
			site.WatchScss(scssPath, configPath) // Следить за scss и пересоздавать css
		}
		h.fiber = fiber.New(fiberConfig)
		h.fiber.Use(recoverMiddleware.New())
		h.fiber.Use(func(c *fiber.Ctx) error {
			c.Locals("config", site.GetConfig(h.configPath))
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
		if config.Routes.Autocomplete != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Autocomplete, config, handlers.Autocomplete)
			} else {
				h.fiber.All(config.Routes.Autocomplete, handlers.Autocomplete)
			}
		}
		if config.Routes.Search != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Search, config, handlers.Search)
			} else {
				h.fiber.All(config.Routes.Search, handlers.Search)
			}
		}
		if config.Routes.Category != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Category, config, handlers.Category)
			} else {
				h.fiber.All(config.Routes.Category, handlers.Category)
			}
		}
		if config.Routes.TopCategories != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.TopCategories, config, handlers.TopCategories)
			} else {
				h.fiber.All(config.Routes.TopCategories, handlers.TopCategories)
			}
		}
		if config.Routes.TopContent != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.TopContent, config, handlers.TopContent)
			} else {
				h.fiber.All(config.Routes.TopContent, handlers.TopContent)
			}
		}
		if config.Routes.Model != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Model, config, handlers.Model)
			} else {
				h.fiber.All(config.Routes.Model, handlers.Model)
			}
		}
		if config.Routes.Channel != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Channel, config, handlers.Channel)
			} else {
				h.fiber.All(config.Routes.Channel, handlers.Channel)
			}
		}
		if config.Routes.ContentItem != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.ContentItem, config, handlers.ContentItem)
			} else {
				h.fiber.All(config.Routes.ContentItem, handlers.ContentItem)
			}
		}
		if config.Routes.New != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.New, config, handlers.New)
			} else {
				h.fiber.All(config.Routes.New, handlers.New)
			}
		}
		if config.Routes.Long != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Long, config, handlers.Long)
			} else {
				h.fiber.All(config.Routes.Long, handlers.Long)
			}
		}
		if config.Routes.Popular != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Popular, config, handlers.Popular)
			} else {
				h.fiber.All(config.Routes.Popular, handlers.Popular)
			}
		}
		if config.Routes.Models != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Models, config, handlers.Models)
			} else {
				h.fiber.All(config.Routes.Models, handlers.Models)
			}
		}
		if config.Routes.Out != "" {
			h.fiber.All(config.Routes.Out, handlers.Out)
		}
		if config.Routes.FakePlayer != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.FakePlayer, config, handlers.FakePlayer)
			} else {
				h.fiber.All(config.Routes.FakePlayer, handlers.FakePlayer)
			}
		}
		if config.Routes.Dmca != "" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(h.fiber, config.Routes.Dmca, config, handlers.Dmca)
			} else {
				h.fiber.All(config.Routes.Dmca, handlers.Dmca)
			}
		}
		if config.Routes.Custom != nil {
			for templateName, routePath := range config.Routes.Custom {
				tName := templateName
				if config.General.MultiLanguage && strings.Contains(routePath, ":lang") {
					handlers.LangHandlers(h.fiber, routePath, config, func(c *fiber.Ctx) error {
						c.Locals("custom_template_name", tName)
						return handlers.Custom(c)
					})
				} else {
					h.fiber.All(routePath, func(c *fiber.Ctx) error {
						c.Locals("custom_template_name", tName)
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
		hostName := strings.TrimPrefix(strings.ToLower(c.Hostname()), "www.")
		host := hosts[hostName]
		if host == nil {
			host = hosts[internal.Config.Frontend.DefaultSite]
		}
		host.fiber.Handler()(c.Context())
		return nil
	})
	return app
}
