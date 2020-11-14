package main

import (
	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"os"
	"path/filepath"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"strings"
)

type host struct {
	fiber  *fiber.App
	config *site.Config
	path   string
}

func InitFiber() *fiber.App {
	fiberConfig := fiber.Config{
		CaseSensitive:         true,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if c.Accepts("text/html") != "" {
				return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
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
		if _, err := toml.DecodeFile(configPath, &h.config); err != nil {
			log.Fatalln("error reading site config at", configPath, err)
		}
		h.fiber = fiber.New(fiberConfig)
		h.fiber.Use(recover.New())
		h.fiber.Use(func(c *fiber.Ctx) error {
			c.Locals("config", h.config)
			c.Locals("path", h.path)
			c.Locals("hostName", hostName)
			return c.Next()
		})
		if h.config.Routes.New != "" {
			h.fiber.Get(h.config.Routes.New, newHandler)
		}
		if h.config.Routes.Autocomplete != "" {
			h.fiber.Get(h.config.Routes.Autocomplete, autocompleteHandler)
		}
		if h.config.Routes.Search != "" {
			h.fiber.Get(h.config.Routes.Search, searchHandler)
		}
		if h.config.Routes.Category != "" {
			h.fiber.Get(h.config.Routes.Category, categoryHandler)
		}
		if h.config.Routes.TopCategories != "" {
			h.fiber.Get(h.config.Routes.TopCategories, topCategoriesHandler)
		}
		if h.config.Routes.TopContent != "" {
			h.fiber.Get(h.config.Routes.TopContent, topContentHandler)
		}
		if h.config.Routes.Model != "" {
			h.fiber.Get(h.config.Routes.Model, modelHandler)
		}
		if h.config.Routes.Channel != "" {
			h.fiber.Get(h.config.Routes.Channel, channelHandler)
		}
		if h.config.Routes.Content != "" {
			h.fiber.Get(h.config.Routes.Content, contentHandler)
		}
		if h.config.Routes.Long != "" {
			h.fiber.Get(h.config.Routes.Long, longHandler)
		}
		if h.config.Routes.Models != "" {
			h.fiber.Get(h.config.Routes.Models, modelsHandler)
		}
		if h.config.Routes.Popular != "" {
			h.fiber.Get(h.config.Routes.Popular, popularHandler)
		}
		if h.config.Routes.Out != "" {
			h.fiber.Get(h.config.Routes.Out, outHandler)
		}
		if h.config.Routes.Custom != nil {
			for templateName, routePath := range h.config.Routes.Custom {
				h.fiber.All(routePath, func(c *fiber.Ctx) error {
					c.Locals("custom_template_name", templateName)
					return c.Next()
				}, customHandler)
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
	app.Use(recover.New())
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
