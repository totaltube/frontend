package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/beevik/etree"
	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var urlRegex = regexp.MustCompile(`^https?://([^/]+)`)

var Sitemap = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	path := r.Context().Value(types.ContextKeyPath).(string)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	currentDate := time.Now().UTC().Format(time.DateOnly)
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if page <= 0 {
		page = 1
	}
	switch r.URL.Query().Get("type") {
	case "main":
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		mainUrls := []string{"top_categories", "top_content", "new", "long", "popular"}
		mainUrls = append(mainUrls, config.Sitemap.AdditionalLinks...)
		for _, uri := range mainUrls {
			link := site.GetLink(uri, config, hostName, config.General.DefaultLanguage, false)
			if link == "" {
				continue
			}
			route := urlSet.CreateElement("url")
			route.CreateElement("loc").
				CreateText("https://" + config.Hostname + link)
			if config.General.MultiLanguage {
				selfAlt := route.CreateElement("xhtml:link")
				selfAlt.CreateAttr("rel", "alternate")
				selfAlt.CreateAttr("hreflang", config.General.DefaultLanguage)
				selfAlt.CreateAttr("href", "https://"+config.Hostname+link)
				xdef := route.CreateElement("xhtml:link")
				xdef.CreateAttr("rel", "alternate")
				xdef.CreateAttr("hreflang", "x-default")
				xdef.CreateAttr("href", "https://"+config.Hostname+link)
				for _, lang := range internal.GetLanguagesAvailableInSitemap(config) {
					if lang.Id == config.General.DefaultLanguage {
						continue
					}
					var hostname = config.Hostname
					if d, ok := config.LanguageDomains[lang.Id]; ok && d != "" {
						matches := urlRegex.FindStringSubmatch(d)
						if len(matches) > 2 {
							hostname = matches[2]
						} else {
							hostname = d
						}
					}
					altLink := site.GetLink(uri, config, hostName, lang.Id, true)
					if altLink != link {
						alt := route.CreateElement("xhtml:link")
						alt.CreateAttr("rel", "alternate")
						alt.CreateAttr("hreflang", lang.Id)
						alt.CreateAttr("href", "https://"+hostname+altLink)
					}
				}
			}
			route.CreateElement("lastmod").CreateText(currentDate)
			route.CreateElement("changefreq").CreateText("hourly")
			route.CreateElement("priority").CreateText("1.0")
		}
	case "categories":
		if config.Sitemap.CategoriesAmount <= 0 || config.Routes.Category == "" || config.Routes.Category == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		pages := (config.Sitemap.CategoriesAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
		var num int64
		if page <= pages {
			results, err := getSitemapCategories(config, config.Hostname, config.Sitemap.CategoriesAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			start := int((page - 1) * config.Sitemap.MaxLinks)
			items := results.Items
			if start < len(items) {
				end := start + int(config.Sitemap.MaxLinks)
				if end > len(items) {
					end = len(items)
				}
				for _, item := range items[start:end] {
					route := urlSet.CreateElement("url")
					route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("category", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
					if config.General.MultiLanguage {
						selfAlt := route.CreateElement("xhtml:link")
						selfAlt.CreateAttr("rel", "alternate")
						selfAlt.CreateAttr("hreflang", config.General.DefaultLanguage)
						selfAlt.CreateAttr("href", "https://"+config.Hostname+site.GetLink("category", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
						xdef := route.CreateElement("xhtml:link")
						xdef.CreateAttr("rel", "alternate")
						xdef.CreateAttr("hreflang", "x-default")
						xdef.CreateAttr("href", "https://"+config.Hostname+site.GetLink("category", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
						for _, lang := range internal.GetLanguagesAvailableInSitemap(config) {
							if lang.Id == config.General.DefaultLanguage {
								continue
							}
							alt := route.CreateElement("xhtml:link")
							alt.CreateAttr("rel", "alternate")
							alt.CreateAttr("hreflang", lang.Id)
							var hostname = config.Hostname
							if d, ok := config.LanguageDomains[lang.Id]; ok && d != "" {
								hostname = d
							}
							alt.CreateAttr("href", "https://"+hostname+site.GetLink("category", config, hostName, lang.Id, true, "slug", item.Slug, "id", item.Id))
						}
					}
					route.CreateElement("lastmod").CreateText(time.Now().UTC().Format(time.DateOnly))
					route.CreateElement("changefreq").CreateText("hourly")
					route.CreateElement("priority").CreateText("0.6")
					num++
					if num >= config.Sitemap.MaxLinks {
						break
					}
				}
			}
		}
	case "models":
		if config.Sitemap.ModelsAmount <= 0 || config.Routes.Model == "" || config.Routes.Model == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		pages := (config.Sitemap.ModelsAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
		var num int64
		if page <= pages {
			results, err := getSitemapModels(config, config.Hostname, config.Sitemap.ModelsAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			start := int((page - 1) * config.Sitemap.MaxLinks)
			items := results.Items
			if start < len(items) {
				end := start + int(config.Sitemap.MaxLinks)
				if end > len(items) {
					end = len(items)
				}
				for _, item := range items[start:end] {
					route := urlSet.CreateElement("url")
					route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("model", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
					if config.General.MultiLanguage {
						selfAlt := route.CreateElement("xhtml:link")
						selfAlt.CreateAttr("rel", "alternate")
						selfAlt.CreateAttr("hreflang", config.General.DefaultLanguage)
						selfAlt.CreateAttr("href", "https://"+config.Hostname+site.GetLink("model", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
						xdef := route.CreateElement("xhtml:link")
						xdef.CreateAttr("rel", "alternate")
						xdef.CreateAttr("hreflang", "x-default")
						xdef.CreateAttr("href", "https://"+config.Hostname+site.GetLink("model", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
						for _, lang := range internal.GetLanguagesAvailableInSitemap(config) {
							if lang.Id == config.General.DefaultLanguage {
								continue
							}
							var hostname = config.Hostname
							if d, ok := config.LanguageDomains[lang.Id]; ok && d != "" {
								hostname = d
							}
							alt := route.CreateElement("xhtml:link")
							alt.CreateAttr("rel", "alternate")
							alt.CreateAttr("hreflang", lang.Id)
							alt.CreateAttr("href", "https://"+hostname+site.GetLink("model", config, hostName, lang.Id, true, "slug", item.Slug, "id", item.Id))
						}
					}
					route.CreateElement("lastmod").CreateText(time.Now().UTC().Format(time.DateOnly))
					route.CreateElement("changefreq").CreateText("hourly")
					route.CreateElement("priority").CreateText("0.6")
					num++
					if num >= config.Sitemap.MaxLinks {
						break
					}
				}
			}
		}
	case "channels":
		if config.Sitemap.ChannelsAmount <= 0 || config.Routes.Channel == "" || config.Routes.Channel == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		pages := (config.Sitemap.ChannelsAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
		var num int64
		if page <= pages {
			results, err := getSitemapChannels(config, config.Hostname, config.Sitemap.ChannelsAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			start := int((page - 1) * config.Sitemap.MaxLinks)
			items := results.Items
			if start < len(items) {
				end := start + int(config.Sitemap.MaxLinks)
				if end > len(items) {
					end = len(items)
				}
				for _, item := range items[start:end] {
					route := urlSet.CreateElement("url")
					route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("channel", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
					if config.General.MultiLanguage {
						selfAlt := route.CreateElement("xhtml:link")
						selfAlt.CreateAttr("rel", "alternate")
						selfAlt.CreateAttr("hreflang", config.General.DefaultLanguage)
						selfAlt.CreateAttr("href", "https://"+config.Hostname+site.GetLink("channel", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
						xdef := route.CreateElement("xhtml:link")
						xdef.CreateAttr("rel", "alternate")
						xdef.CreateAttr("hreflang", "x-default")
						xdef.CreateAttr("href", "https://"+config.Hostname+site.GetLink("channel", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
						for _, lang := range internal.GetLanguagesAvailableInSitemap(config) {
							if lang.Id == config.General.DefaultLanguage {
								continue
							}
							var hostname = config.Hostname
							if d, ok := config.LanguageDomains[lang.Id]; ok && d != "" {
								hostname = d
							}
							alt := route.CreateElement("xhtml:link")
							alt.CreateAttr("rel", "alternate")
							alt.CreateAttr("hreflang", lang.Id)
							alt.CreateAttr("href", "https://"+hostname+site.GetLink("channel", config, hostName, lang.Id, true, "slug", item.Slug, "id", item.Id))
						}
					}
					route.CreateElement("lastmod").CreateText(time.Now().UTC().Format(time.DateOnly))
					route.CreateElement("changefreq").CreateText("hourly")
					route.CreateElement("priority").CreateText("0.6")
					num++
					if num >= config.Sitemap.MaxLinks {
						break
					}
				}
			}
		}
	case "videos":
		if config.Sitemap.LastVideosAmount <= 0 || config.Routes.ContentItem == "" || config.Routes.ContentItem == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		pages := (config.Sitemap.LastVideosAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
		var num int64
		if page <= pages {
			results, err := getSitemapVideos(config, config.Hostname, config.Sitemap.MaxLinks, page)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			customContext := generateCustomContext(w, r, "sitemap-video")
			for _, item := range results.Items {
				var videoBytes []byte
				videoBytes, err = site.ParseTemplate("sitemap-video", path, config, customContext, true, fmt.Sprintf("sitemap-video-%d", item.Id), 1, func() (ctx pongo2.Context, err error) {
					ctx = pongo2.Context{
						"content_item": item,
					}
					return
				}, w, r)
				if err == nil {
					// получаем из шаблона url для видео
					video := etree.NewDocument()
					err = video.ReadFromBytes(videoBytes)
					if err != nil {
						log.Println(err, string(videoBytes))
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					urlSet.AddChild(video.Root())
					num++
					if num >= config.Sitemap.MaxLinks {
						break
					}
					continue
				}
				if !errors.Is(err, site.ErrTemplateNotFound) {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				route := urlSet.CreateElement("url")
				route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink(
					"content_item", config, hostName, config.General.DefaultLanguage, false,
					"slug", item.Slug, "id", item.Id, "categories", item.Categories))
				if config.General.MultiLanguage {
					selfAlt := route.CreateElement("xhtml:link")
					selfAlt.CreateAttr("rel", "alternate")
					selfAlt.CreateAttr("hreflang", config.General.DefaultLanguage)
					selfAlt.CreateAttr("href", "https://"+config.Hostname+site.GetLink(
						"content_item", config, hostName, config.General.DefaultLanguage, false,
						"slug", item.Slug, "id", item.Id, "categories", item.Categories))
					xdef := route.CreateElement("xhtml:link")
					xdef.CreateAttr("rel", "alternate")
					xdef.CreateAttr("hreflang", "x-default")
					xdef.CreateAttr("href", "https://"+config.Hostname+site.GetLink(
						"content_item", config, hostName, config.General.DefaultLanguage, false,
						"slug", item.Slug, "id", item.Id, "categories", item.Categories))
					for _, lang := range internal.GetLanguagesAvailableInSitemap(config) {
						if lang.Id == config.General.DefaultLanguage {
							continue
						}
						var hostname = config.Hostname
						if d, ok := config.LanguageDomains[lang.Id]; ok && d != "" {
							hostname = d
						}
						alt := route.CreateElement("xhtml:link")
						alt.CreateAttr("rel", "alternate")
						alt.CreateAttr("hreflang", lang.Id)
						alt.CreateAttr("href", "https://"+hostname+site.GetLink(
							"content_item", config, hostName, lang.Id, true,
							"slug", item.Slug, "id", item.Id, "categories", item.Categories))
					}
				}
				route.CreateElement("lastmod").CreateText(time.Now().UTC().Format(time.DateOnly))
				route.CreateElement("changefreq").CreateText("daily")
				route.CreateElement("priority").CreateText("0.8")
				/*
					video := route.CreateElement("video:video")
					video.CreateElement("video:thumbnail_loc").CreateText(item.HiresThumb())
					video.CreateElement("video:title").CreateCData(item.Title)
					if item.Description != nil && *item.Description != "" {
						video.CreateElement("video:description").CreateCData(*item.Description)
					}
					video.CreateElement("video:view_count").CreateText(strconv.FormatInt(int64(item.Views), 10))
					video.CreateElement("video:publication_date").CreateText(item.Dated.Format(time.DateOnly))
					video.CreateElement("video:duration").CreateText(fmt.Sprintf("%d", item.Duration))
					for _, cat := range item.Categories {
						video.CreateElement("video:category").CreateCData(cat.Title)
					}
				*/
				num++
				if num >= config.Sitemap.MaxLinks {
					break
				}
			}
		}
	case "searches":
		if config.Sitemap.SearchesAmount <= 0 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		lang := r.URL.Query().Get("lang")
		if lang == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		pages := (config.Sitemap.SearchesAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
		var num int64
		if page <= pages {
			results, err := getSitemapSearches(config, config.Hostname, lang, config.Sitemap.SearchesAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			start := int((page - 1) * config.Sitemap.MaxLinks)
			if start < len(results) {
				end := start + int(config.Sitemap.MaxLinks)
				if end > len(results) {
					end = len(results)
				}
				for _, item := range results[start:end] {
					route := urlSet.CreateElement("url")
					route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("search", config, hostName, lang, false, "query", item.Message))
					route.CreateElement("lastmod").CreateText(time.Now().UTC().Format(time.DateOnly))
					route.CreateElement("changefreq").CreateText("hourly")
					route.CreateElement("priority").CreateText("0.6")
					num++
					if num >= config.Sitemap.MaxLinks {
						break
					}
				}
			}
		}
	default:
		// Showing sitemap index
		sitemapIndex := doc.CreateElement("sitemapindex")
		sitemapIndex.CreateAttr("xmlns", `http://www.sitemaps.org/schemas/sitemap/0.9`)
		main := sitemapIndex.CreateElement("sitemap")
		main.CreateElement("loc").CreateText("https://" + config.Hostname + config.Sitemap.Route + "?type=main")
		main.CreateElement("lastmod").CreateText(currentDate)
		if config.Sitemap.CategoriesAmount > 0 && config.Routes.Category != "" && config.Routes.Category != "-" {
			pages := (config.Sitemap.CategoriesAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
			for i := int64(0); i < pages; i++ {
				categories := sitemapIndex.CreateElement("sitemap")
				pageStr := ""
				if i > 0 {
					pageStr = fmt.Sprintf("&page=%d", i+1)
				}
				categories.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=categories%s", config.Hostname, config.Sitemap.Route, pageStr),
				)
				categories.CreateElement("lastmod").CreateText(currentDate)
			}
		}
		if config.Sitemap.ModelsAmount > 0 && config.Routes.Model != "" && config.Routes.Model != "-" {
			pages := (config.Sitemap.ModelsAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
			for i := int64(0); i < pages; i++ {
				models := sitemapIndex.CreateElement("sitemap")
				pageStr := ""
				if i > 0 {
					pageStr = fmt.Sprintf("&page=%d", i+1)
				}
				models.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=models%s", config.Hostname, config.Sitemap.Route, pageStr),
				)
				models.CreateElement("lastmod").CreateText(currentDate)
			}
		}
		if config.Sitemap.ChannelsAmount > 0 && config.Routes.Channel != "" && config.Routes.Channel != "-" {
			pages := (config.Sitemap.ChannelsAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
			for i := int64(0); i < pages; i++ {
				channels := sitemapIndex.CreateElement("sitemap")
				pageStr := ""
				if i > 0 {
					pageStr = fmt.Sprintf("&page=%d", i+1)
				}
				channels.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=channels%s", config.Hostname, config.Sitemap.Route, pageStr),
				)
				channels.CreateElement("lastmod").CreateText(currentDate)
			}
		}
		if config.Sitemap.LastVideosAmount > 0 && config.Routes.ContentItem != "" && config.Routes.ContentItem != "-" {
			pages := (config.Sitemap.LastVideosAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
			for i := int64(0); i < pages; i++ {
				videos := sitemapIndex.CreateElement("sitemap")
				pageStr := ""
				if i > 0 {
					pageStr = fmt.Sprintf("&page=%d", i+1)
				}
				videos.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=videos%s", config.Hostname, config.Sitemap.Route, pageStr),
				)
				videos.CreateElement("lastmod").CreateText(currentDate)
			}
		}
		if config.Sitemap.SearchesAmount > 0 && config.Routes.Search != "" && config.Routes.Search != "-" {
			langs := []string{config.General.DefaultLanguage}
			if config.General.MultiLanguage {
				langs = lo.Map(internal.GetLanguagesAvailableInSitemap(config), func(t types.Language, i int) string {
					return t.Id
				})
			}
			pages := (config.Sitemap.SearchesAmount + config.Sitemap.MaxLinks - 1) / config.Sitemap.MaxLinks
			for _, lang := range langs {
				for i := int64(0); i < pages; i++ {
					searches := sitemapIndex.CreateElement("sitemap")
					pageStr := ""
					if i > 0 {
						pageStr = fmt.Sprintf("&page=%d", i+1)
					}
					searches.CreateElement("loc").CreateText(
						fmt.Sprintf("https://%s%s?type=searches&lang=%s%s", config.Hostname, config.Sitemap.Route, url.QueryEscape(lang), pageStr),
					)
					searches.CreateElement("lastmod").CreateText(currentDate)
				}
			}
		}
	}
	// doc.Indent(2) // pretty-printing XML is CPU-expensive and not needed for sitemaps
	if middlewares.HeadersSent(w) {
		return
	}
	w.Header().Add("Content-Type", "application/xml; charset=utf-8")
	w.Header().Add("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate")
	w.Header().Add("Expires", time.Now().Add(-time.Hour*24).Format(http.TimeFormat))
	w.Header().Add("Pragma", "no-cache")
	_, _ = doc.WriteTo(w)
})

func getSitemapCategories(siteConfig *types.Config, hostName string, amount int64) (results *types.CategoryResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-categories-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.CategoriesList(siteConfig, "en", 1, api.SortPopular, amount, 0)
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapModels(siteConfig *types.Config, hostName string, amount int64) (results *types.ModelResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-models-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		rawResponse, err = api.ModelsListRaw(siteConfig, "en", 1, api.SortPopular, amount, "", 0)
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.ModelResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapChannels(siteConfig *types.Config, hostName string, amount int64) (results *types.ChannelResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-channels-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.ChannelsList(siteConfig, "en", 1, api.SortPopular, amount, 0)
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.ChannelResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapVideos(siteConfig *types.Config, hostName string, amount int64, page int64) (results *types.ContentResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-videos-%d-%d", hostName, amount, page), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		rawResponse, err = api.ContentRaw(siteConfig, api.ContentParams{
			Amount: amount,
			Sort:   api.SortDated,
			Page:   page,
		})
		return rawResponse, err
	}, false); err != nil {
		log.Println(err)
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(cached, results)
	if err != nil {
		log.Println(err)
	}
	return
}

func getSitemapSearches(siteConfig *types.Config, hostName string, lang string, amount int64) (results []types.TopSearch, err error) {
	var ttl = time.Hour + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:%s:top-searches-%d", hostName, lang, amount), ttl, ttl, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.TopSearches(siteConfig, lang, amount)
		return rawResponse, err
	}, false); err != nil {
		log.Println(err)
		return
	}
	result := struct {
		Items []types.TopSearch `json:"items"`
	}{}
	err = json.Unmarshal(cached, &result)
	if err != nil {
		log.Println(err)
		return
	}
	results = result.Items
	return
}
