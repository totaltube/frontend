package main

import (
	"context"
	"github.com/samber/lo"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sersh.com/totaltube/frontend/handlers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/site"
)

type hostRouter struct {
	configPath string
	path       string
	handler    http.Handler
}

func InitRouter() http.Handler {
	hosts := map[string]*hostRouter{}
	defaultSiteOk := false
	matches, err := filepath.Glob(filepath.Join(internal.Config.Frontend.SitesPath, "*"))
	if err != nil {
		log.Fatalln(err)
	}
	for _, m := range matches {
		m := m
		configPath := filepath.Join(m, "config.toml")
		if _, err := os.Stat(configPath); err != nil {
			continue
		}
		config := internal.GetConfigAndWatch(configPath)
		h := hostRouter{
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
		hr := chi.NewRouter()
		hr.Use(middleware.Recoverer)
		hr.Use(middleware.Timeout(60 * time.Second))
		hr.Use(middleware.GetHead)
		if internal.Config.General.EnableAccessLog {
			hr.Use(middleware.Logger)
		}
		hr.Use(middleware.StripSlashes)
		hr.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), "config", internal.GetConfig(configPath))
				ctx = context.WithValue(ctx, "path", h.path)
				ctx = context.WithValue(ctx, "hostName", hostName)
				ctx = context.WithValue(ctx, "lang", "en")
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		dir := http.Dir(filepath.Join(m, "public"))
		fileServer := http.FileServer(dir)
		// Serving static if it exists in the public route
		hr.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if (r.Method == "GET" || r.Method == "") && strings.ContainsRune(r.URL.Path, '.') {
					if _, err := os.Stat(filepath.Join(m, "public", r.URL.Path)); err == nil {
						fileServer.ServeHTTP(w, r)
						return
					}
				}
				next.ServeHTTP(w, r)
			})
		})
		if config.Routes.Rating != "" && config.Routes.Rating != "-" {
			hr.Handle(config.Routes.Rating, middlewares.BadBotMiddleware(handlers.Rating))
		}
		if config.Routes.Autocomplete != "" && config.Routes.Autocomplete != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Autocomplete, config, handlers.Autocomplete)
			} else {
				hr.Handle(config.Routes.Autocomplete, middlewares.BadBotMiddleware(handlers.Autocomplete))
			}
		}
		if config.Routes.Search != "" && config.Routes.Search != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Search, config, handlers.Search)
			} else {
				hr.Handle(config.Routes.Search, middlewares.BadBotMiddleware(handlers.Search))
			}
		}
		if config.Routes.Category != "" && config.Routes.Category != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Category, config, handlers.Category)
			} else {
				hr.Handle(config.Routes.Category, middlewares.BadBotMiddleware(handlers.Category))
			}
		}
		if config.Routes.TopCategories != "" && config.Routes.TopCategories != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.TopCategories, config, handlers.TopCategories)
			} else {
				hr.Handle(config.Routes.TopCategories, middlewares.BadBotMiddleware(handlers.TopCategories))
			}
		}
		if config.Routes.TopContent != "" && config.Routes.TopContent != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.TopContent, config, handlers.TopContent)
			} else {
				hr.Handle(config.Routes.TopContent, middlewares.BadBotMiddleware(handlers.TopContent))
			}
		}
		if config.Routes.Model != "" && config.Routes.Model != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Model, config, handlers.Model)
			} else {
				hr.Handle(config.Routes.Model, middlewares.BadBotMiddleware(handlers.Model))
			}
		}
		if config.Routes.Channel != "" && config.Routes.Channel != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Channel, config, handlers.Channel)
			} else {
				hr.Handle(config.Routes.Channel, middlewares.BadBotMiddleware(handlers.Channel))
			}
		}
		if config.Routes.ContentItem != "" && config.Routes.ContentItem != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.ContentItem, config, handlers.ContentItem)
			} else {
				hr.Handle(config.Routes.ContentItem, middlewares.BadBotMiddleware(handlers.ContentItem))
			}
		}
		if config.Routes.New != "" && config.Routes.New != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.New, config, handlers.New)
			} else {
				hr.Handle(config.Routes.New, middlewares.BadBotMiddleware(handlers.New))
			}
		}
		if config.Routes.Long != "" && config.Routes.Long != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Long, config, handlers.Long)
			} else {
				hr.Handle(config.Routes.Long, middlewares.BadBotMiddleware(handlers.Long))
			}
		}
		if config.Routes.Popular != "" && config.Routes.Popular != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Popular, config, handlers.Popular)
			} else {
				hr.Handle(config.Routes.Popular, middlewares.BadBotMiddleware(handlers.Popular))
			}
		}
		if config.Routes.Models != "" && config.Routes.Models != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Models, config, handlers.Models)
			} else {
				hr.Handle(config.Routes.Models, middlewares.BadBotMiddleware(handlers.Models))
			}
		}
		if config.Routes.Out != "" && config.Routes.Out != "-" {
			hr.Handle(config.Routes.Out, middlewares.BadBotMiddleware(handlers.Out))
		}
		if config.Routes.FakePlayer != "" && config.Routes.FakePlayer != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.FakePlayer, config, handlers.FakePlayer)
			} else {
				hr.Handle(config.Routes.FakePlayer, middlewares.BadBotMiddleware(handlers.FakePlayer))
			}
		}
		if config.Routes.VideoEmbed != "" && config.Routes.VideoEmbed != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.VideoEmbed, config, handlers.VideoEmbed)
			} else {
				hr.Handle(config.Routes.VideoEmbed, middlewares.BadBotMiddleware(handlers.VideoEmbed))
			}
		}
		if config.Routes.Dmca != "" && config.Routes.Dmca != "-" {
			if config.General.MultiLanguage {
				handlers.LangHandlers(hr, config.Routes.Dmca, config, handlers.Dmca)
			} else {
				hr.Handle(config.Routes.Dmca, middlewares.BadBotMiddleware(handlers.Dmca))
			}
		}
		if internal.Config.Frontend.RouteRedirectContentItem != "" && internal.Config.Frontend.RouteRedirectContentItem != "-" {
			hr.Handle(internal.Config.Frontend.RouteRedirectContentItem, handlers.RedirectToContentItem)
		}
		if config.Routes.Custom != nil {
			for templateName, routePath := range config.Routes.Custom {
				if strings.HasSuffix(templateName, "_multilang") {
					continue
				}
				tName := templateName
				_, isCustomMultilangTemplate := config.Routes.Custom[tName+"_multilang"]
				if config.General.MultiLanguage && (strings.Contains(routePath, "{lang}") || isCustomMultilangTemplate) {
					handlers.LangHandlers(hr, routePath, config, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						ctx := context.WithValue(r.Context(), "custom_template_name", tName)
						middlewares.BadBotMiddleware(handlers.Custom).ServeHTTP(w, r.WithContext(ctx))
					}))
				} else {
					hr.Handle(routePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						ctx := context.WithValue(r.Context(), "custom_template_name", tName)
						middlewares.BadBotMiddleware(handlers.Custom).ServeHTTP(w, r.WithContext(ctx))
					}))
				}
			}
		}
		if config.General.ToplistDataUrl == "" {
			config.General.ToplistDataUrl = internal.Config.General.ToplistDataUrl
		}
		if config.General.ToplistDataUrl != "" {
			hr.Handle(config.General.ToplistDataUrl, middlewares.BadBotMiddleware(handlers.ToplistData))
		}
		if config.Sitemap.Route != "" {
			hr.Handle(config.Sitemap.Route, handlers.Sitemap)
		}
		blackholeRoute := internal.Config.General.DefaultBlackholeRoute
		if config.Routes.Blackhole != "" {
			blackholeRoute = config.Routes.Blackhole
		}
		if blackholeRoute != "" && blackholeRoute != "-" {
			lo.ForEach(strings.Split(blackholeRoute, ","), func(s string, i int) {
				route := strings.TrimSpace(s)
				hr.Handle(route, handlers.Blackhole)
			})
		}
		hr.NotFound(handlers.Handle404)
		//hr.Handle("/*", handlers.Handle404)
		h.handler = hr
		hosts[hostName] = &h
		if hostName == internal.Config.Frontend.DefaultSite {
			defaultSiteOk = true
		}
	}
	if !defaultSiteOk {
		log.Fatalln("Default site", internal.Config.Frontend.DefaultSite,
			"not present in", internal.Config.Frontend.SitesPath)
	}
	r := chi.NewRouter()
	if internal.Config.General.Development {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Use(middlewares.Timeout(10 * time.Second))
	r.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostName := strings.TrimPrefix(strings.ToLower(r.Host), "www.")
		host := hosts[hostName]
		if host == nil {
			log.Println("can't find hostname", hostName, ", defaulting to", internal.Config.Frontend.DefaultSite)
			host = hosts[internal.Config.Frontend.DefaultSite]
		}
		// get first not empty value
		ip := r.Header.Get(internal.Config.General.RealIpHeader)
		if ip == "" {
			ip = r.RemoteAddr
		}
		if ip == "" || len(ip) < 7 {
			ip = "127.0.0.1"
		}
		ctx := context.WithValue(r.Context(), "ip", ip)
		host.handler.ServeHTTP(w, r.WithContext(ctx))
		return
	}))
	if os.Getenv("GO_ENV") == "debug" {
		r.Mount("/_debug", middleware.Profiler())
	}
	return r
}
