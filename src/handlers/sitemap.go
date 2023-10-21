package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"github.com/flosch/pongo2/v4"
	"github.com/samber/lo"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"time"
)

var Sitemap = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*types.Config)
	path := r.Context().Value("path").(string)
	hostName := r.Context().Value("hostName").(string)
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
		for _, c := range config.Sitemap.AdditionalLinks {
			mainUrls = append(mainUrls, c)
		}
		for _, uri := range mainUrls {
			link := site.GetLink(uri, config, hostName, config.General.DefaultLanguage, false)
			if link == "" {
				continue
			}
			route := urlSet.CreateElement("url")
			route.CreateElement("loc").
				CreateText("https://" + config.Hostname + link)
			if config.General.MultiLanguage {
				for _, lang := range internal.GetLanguages() {
					altLink := site.GetLink(uri, config, hostName, lang.Id, true)
					if altLink != link {
						alt := route.CreateElement("xhtml:link")
						alt.CreateAttr("rel", "alternate")
						alt.CreateAttr("hreflang", lang.Id)
						alt.CreateAttr("href", "https://"+config.Hostname+altLink)
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
		results, err := getSitemapCategories(config.Hostname, config.Sitemap.CategoriesAmount)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if int64(len(results.Items)) <= (page-1)*config.Sitemap.MaxLinks {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var num int64
		for _, item := range results.Items[(page-1)*config.Sitemap.MaxLinks:] {
			route := urlSet.CreateElement("url")
			route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("category", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
			if config.General.MultiLanguage {
				for _, lang := range internal.GetLanguages() {
					alt := route.CreateElement("xhtml:link")
					alt.CreateAttr("rel", "alternate")
					alt.CreateAttr("hreflang", lang.Id)
					alt.CreateAttr("href", "https://"+config.Hostname+site.GetLink("category", config, hostName, lang.Id, true, "slug", item.Slug, "id", item.Id))
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
	case "models":
		if config.Sitemap.ModelsAmount <= 0 || config.Routes.Model == "" || config.Routes.Model == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		results, err := getSitemapModels(config.Hostname, config.Sitemap.ModelsAmount)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if int64(len(results.Items)) <= (page-1)*config.Sitemap.MaxLinks {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var num int64
		for _, item := range results.Items[(page-1)*config.Sitemap.MaxLinks:] {
			route := urlSet.CreateElement("url")
			route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("model", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
			if config.General.MultiLanguage {
				for _, lang := range internal.GetLanguages() {
					alt := route.CreateElement("xhtml:link")
					alt.CreateAttr("rel", "alternate")
					alt.CreateAttr("hreflang", lang.Id)
					alt.CreateAttr("href", "https://"+config.Hostname+site.GetLink("model", config, hostName, lang.Id, true, "slug", item.Slug, "id", item.Id))
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
	case "channels":
		if config.Sitemap.ChannelsAmount <= 0 || config.Routes.Channel == "" || config.Routes.Channel == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		results, err := getSitemapChannels(config.Hostname, config.Sitemap.ChannelsAmount)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if int64(len(results.Items)) <= (page-1)*config.Sitemap.MaxLinks {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var num int64
		for _, item := range results.Items[(page-1)*config.Sitemap.MaxLinks:] {
			route := urlSet.CreateElement("url")
			route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink("channel", config, hostName, config.General.DefaultLanguage, false, "slug", item.Slug, "id", item.Id))
			if config.General.MultiLanguage {
				for _, lang := range internal.GetLanguages() {
					alt := route.CreateElement("xhtml:link")
					alt.CreateAttr("rel", "alternate")
					alt.CreateAttr("hreflang", lang.Id)
					alt.CreateAttr("href", "https://"+config.Hostname+site.GetLink("channel", config, hostName, lang.Id, true, "slug", item.Slug, "id", item.Id))
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
	case "videos":
		if config.Sitemap.ChannelsAmount <= 0 || config.Routes.ContentItem == "" || config.Routes.ContentItem == "-" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		urlSet := doc.CreateElement("urlset")
		urlSet.CreateAttr("xmlns", "http://www.sitemaps.org/schemas/sitemap/0.9")
		urlSet.CreateAttr("xmlns:video", `http://www.google.com/schemas/sitemap-video/1.1`)
		urlSet.CreateAttr("xmlns:xhtml", `http://www.w3.org/1999/xhtml`)
		results, err := getSitemapVideos(config.Hostname, config.Sitemap.LastVideosAmount)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if int64(len(results.Items)) <= (page-1)*config.Sitemap.MaxLinks {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var num int64
		customContext := generateCustomContext(w, r, "sitemap-video")
		for _, item := range results.Items[(page-1)*config.Sitemap.MaxLinks:] {
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
			if err != site.ErrTemplateNotFound {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			route := urlSet.CreateElement("url")
			route.CreateElement("loc").CreateText("https://" + config.Hostname + site.GetLink(
				"content_item", config, hostName, config.General.DefaultLanguage, false,
				"slug", item.Slug, "id", item.Id, "categories", item.Categories))
			if config.General.MultiLanguage {
				for _, lang := range internal.GetLanguages() {
					alt := route.CreateElement("xhtml:link")
					alt.CreateAttr("rel", "alternate")
					alt.CreateAttr("hreflang", lang.Id)
					alt.CreateAttr("href", "https://"+config.Hostname+site.GetLink(
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
		results, err := getSitemapSearches(config.Hostname, lang, config.Sitemap.SearchesAmount)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if int64(len(results)) <= (page-1)*config.Sitemap.MaxLinks {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		var num int64
		for _, item := range results[(page-1)*config.Sitemap.MaxLinks:] {
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
	default:
		// Showing sitemap index
		sitemapIndex := doc.CreateElement("sitemapindex")
		sitemapIndex.CreateAttr("xmlns", `http://www.sitemaps.org/schemas/sitemap/0.9`)
		main := sitemapIndex.CreateElement("sitemap")
		main.CreateElement("loc").CreateText("https://" + config.Hostname + config.Sitemap.Route + "?type=main")
		main.CreateElement("lastmod").CreateText(currentDate)
		if config.Sitemap.CategoriesAmount > 0 && config.Routes.Category != "" && config.Routes.Category != "-" {
			results, err := getSitemapCategories(config.Hostname, config.Sitemap.CategoriesAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			maxAmount := int64(len(results.Items))
			var lastFrom int64
			for lastFrom < maxAmount {
				categories := sitemapIndex.CreateElement("sitemap")
				page := ""
				if lastFrom > 0 {
					page = fmt.Sprintf("&page=%d", lastFrom/config.Sitemap.MaxLinks+1)
				}
				categories.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=categories%s", config.Hostname, config.Sitemap.Route, page),
				)
				categories.CreateElement("lastmod").CreateText(currentDate)
				lastFrom += config.Sitemap.MaxLinks
			}
		}
		if config.Sitemap.ModelsAmount > 0 && config.Routes.Model != "" && config.Routes.Model != "-" {
			results, err := getSitemapModels(config.Hostname, config.Sitemap.ModelsAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			maxAmount := int64(len(results.Items))
			var lastFrom int64
			for lastFrom < maxAmount {
				models := sitemapIndex.CreateElement("sitemap")
				page := ""
				if lastFrom > 0 {
					page = fmt.Sprintf("&page=%d", lastFrom/config.Sitemap.MaxLinks+1)
				}
				models.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=models%s", config.Hostname, config.Sitemap.Route, page),
				)
				models.CreateElement("lastmod").CreateText(currentDate)
				lastFrom += config.Sitemap.MaxLinks
			}
		}
		if config.Sitemap.ChannelsAmount > 0 && config.Routes.Channel != "" && config.Routes.Channel != "-" {
			results, err := getSitemapChannels(config.Hostname, config.Sitemap.CategoriesAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			maxAmount := int64(len(results.Items))
			var lastFrom int64
			for lastFrom < maxAmount {
				models := sitemapIndex.CreateElement("sitemap")
				page := ""
				if lastFrom > 0 {
					page = fmt.Sprintf("&page=%d", lastFrom/config.Sitemap.MaxLinks+1)
				}
				models.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=channels%s", config.Hostname, config.Sitemap.Route, page),
				)
				models.CreateElement("lastmod").CreateText(currentDate)
				lastFrom += config.Sitemap.MaxLinks
			}
		}
		if config.Sitemap.LastVideosAmount > 0 && config.Routes.ContentItem != "" && config.Routes.ContentItem != "-" {
			results, err := getSitemapVideos(config.Hostname, config.Sitemap.LastVideosAmount)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			maxAmount := int64(len(results.Items))
			var lastFrom int64
			for lastFrom < maxAmount {
				videos := sitemapIndex.CreateElement("sitemap")
				page := ""
				if lastFrom > 0 {
					page = fmt.Sprintf("&page=%d", lastFrom/config.Sitemap.MaxLinks+1)
				}
				videos.CreateElement("loc").CreateText(
					fmt.Sprintf("https://%s%s?type=videos%s", config.Hostname, config.Sitemap.Route, page),
				)
				videos.CreateElement("lastmod").CreateText(currentDate)
				lastFrom += config.Sitemap.MaxLinks
			}
		}
		if config.Sitemap.SearchesAmount > 0 && config.Routes.Search != "" && config.Routes.Search != "-" {
			langs := []string{config.General.DefaultLanguage}
			if config.General.MultiLanguage {
				langs = lo.Map(internal.GetLanguages(), func(t types.Language, i int) string {
					return t.Id
				})
			}
			for _, lang := range langs {
				results, err := getSitemapSearches(config.Hostname, lang, config.Sitemap.SearchesAmount)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				maxAmount := int64(len(results))
				var lastFrom int64
				for lastFrom < maxAmount {
					searches := sitemapIndex.CreateElement("sitemap")
					page := ""
					if lastFrom > 0 {
						page = fmt.Sprintf("&page=%d", lastFrom/config.Sitemap.MaxLinks+1)
					}
					searches.CreateElement("loc").CreateText(
						fmt.Sprintf("https://%s%s?type=searches&lang=%s%s", config.Hostname, config.Sitemap.Route, url.QueryEscape(lang), page),
					)
					searches.CreateElement("lastmod").CreateText(currentDate)
					lastFrom += config.Sitemap.MaxLinks
				}
			}
		}
	}
	doc.Indent(2)
	w.Header().Add("Content-Type", "application/xml; charset=utf-8")
	w.Header().Add("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate")
	w.Header().Add("Expires", time.Now().Add(-time.Hour*24).Format(http.TimeFormat))
	w.Header().Add("Pragma", "no-cache")
	_, _ = doc.WriteTo(w)
})

func getSitemapCategories(hostName string, amount int64) (results *types.CategoryResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-categories-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.CategoriesList(hostName, "en", 1, api.SortPopular, amount, 0)
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapModels(hostName string, amount int64) (results *types.ModelResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-models-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		rawResponse, err = api.ModelsListRaw(hostName, "en", 1, api.SortPopular, amount, "", 0)
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.ModelResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapChannels(hostName string, amount int64) (results *types.ChannelResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-channels-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.ChannelsList(hostName, "en", 1, api.SortPopular, amount, 0)
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.ChannelResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapVideos(hostName string, amount int64) (results *types.ContentResults, err error) {
	var ttl = time.Hour*2 + time.Duration(rand.Intn(3600))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:top-videos-%d", hostName, amount), ttl, time.Hour*2, func() ([]byte, error) {
		var rawResponse json.RawMessage
		rawResponse, err = api.ContentRaw(hostName, api.ContentParams{
			Amount:    amount,
			Sort:      api.SortDated,
			Timeframe: "",
		})
		return rawResponse, err
	}, false); err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(cached, results)
	return
}

func getSitemapSearches(hostName string, lang string, amount int64) (results []types.TopSearch, err error) {
	var ttl = time.Hour*24 + time.Duration(rand.Intn(3600*10))*time.Second
	var cached []byte
	if cached, err = db.GetCachedTimeout(fmt.Sprintf("sitemap:%s:%s:top-searches-%d", hostName, lang, amount), ttl, time.Hour*20, func() ([]byte, error) {
		var rawResponse json.RawMessage
		_, rawResponse, err = api.TopSearches(hostName, lang, amount)
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
