package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

func getTopCategoriesFunc(config *types.Config, langId string, groupId int64) func(args ...interface{}) *types.CategoryResults {
	return func(args ...interface{}) *types.CategoryResults {
		parsingName := true
		var page int64 = 1
		curName := ""
		for k := range args {
			if parsingName {
				curName = fmt.Sprintf("%v", args[k])
				parsingName = false
				continue
			}
			val := fmt.Sprintf("%v", args[k])
			parsingName = true
			switch curName {
			case "lang":
				langId = val
			case "page":
				page, _ = strconv.ParseInt(val, 10, 32)
			case "group_id":
				groupId, _ = strconv.ParseInt(val, 10, 32)
			}
		}
		results, err := api.TopCategories(config, langId, page, groupId)
		if err != nil {
			log.Println("can't get top categories:", err)
			return nil
		}
		if page == 1 {
			randomizeRatio := config.General.RandomizeRatio
			if randomizeRatio < 0 {
				randomizeRatio = internal.Config.General.RandomizeRatio
			}
			if randomizeRatio > 0 {
				helpers.RandomizeItems(results.Items, randomizeRatio)
			}
		}
		return results
	}
}

var TopCategories = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	langId := r.Context().Value(types.ContextKeyLang).(string)
	ip := r.Context().Value(types.ContextKeyIp).(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	if ref := r.Header.Get("Referer"); ref != "" && page == 1 && !config.General.DisableCategoriesRedirect {
		if u, err := url.Parse(ref); err == nil &&
			strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") != hostName &&
			!botDetector.IsBot(r.Header.Get("User-Agent")) {
			var s = strings.ToLower(u.Path + " " + u.RawQuery)
			if categories, err := db.GetCachedTopCategories(config, hostName, groupId); err == nil {
				for _, cat := range categories.Items {
					for _, t := range cat.Tags {
						if strings.Contains(s, t) {
							var redirectUrl = config.Routes.Category
							if config.General.MultiLanguage {
								redirectUrl = strings.ReplaceAll(config.Routes.LanguageTemplate, "{lang}", langId)
								redirectUrl = strings.ReplaceAll(redirectUrl, "{route}", config.Routes.Category)
							}
							redirectUrl = strings.ReplaceAll(redirectUrl, "{slug}", cat.Slug)
							redirectUrl = strings.ReplaceAll(redirectUrl, "{id}", strconv.FormatInt(int64(cat.Id), 10))
							if qs := r.URL.RawQuery; qs != "" {
								redirectUrl = redirectUrl + "?" + qs
							}
							http.Redirect(w, r, redirectUrl, http.StatusFound)
							if internal.Config.General.EnableAccessLog {
								log.Println(hostName, 302, redirectUrl)
							}
							return
						}
					}
				}
			}
		}
	}
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	customContext := generateCustomContext(w, r, "top-categories")
	cacheKey := fmt.Sprintf("top-categories:%s:%s:%d:%d", hostName, langId, page, groupId)
	var cacheTtl types.Duration
	if config.CacheTimeouts.TopCategories != nil {
		cacheTtl = *config.CacheTimeouts.TopCategories
	} else {
		cacheTtl = internal.Config.CacheTimeouts.TopCategories
	}
	if page > 1 {
		if config.CacheTimeouts.TopCategoriesPagination != nil {
			cacheTtl = *config.CacheTimeouts.TopCategoriesPagination
		} else {
			cacheTtl = internal.Config.CacheTimeouts.TopCategoriesPagination
		}
	}
	pageTtl := 0 * time.Second
	randomizeRatio := config.General.RandomizeRatio
	if randomizeRatio < 0 {
		randomizeRatio = internal.Config.General.RandomizeRatio
	}
	if randomizeRatio <= 0 {
		pageTtl = time.Second * 15
	}
	parsed, err := site.ParseTemplate("top-categories", path, config, customContext, nocache, cacheKey, pageTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			var results = new(types.CategoryResults)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", time.Duration(cacheTtl), time.Duration(cacheTtl), func() ([]byte, error) {
				bt, err := api.TopCategoriesRaw(config, langId, page, groupId)
				return bt, err
			}, nocache)
			if err != nil {
				return ctx, err
			}
			err = json.Unmarshal(response, results)
			if err != nil {
				return ctx, err
			}
			if len(results.Items) == 0 && page > 1 {
				return ctx, fmt.Errorf("not found")
			}
			if page == 1 && randomizeRatio > 0 {
				helpers.RandomizeItems(results.Items, randomizeRatio)
			}
			ctx["top_categories"] = results
			ctx["total"] = int64(results.Total)
			ctx["from"] = int64(results.From)
			ctx["to"] = int64(results.To)
			ctx["page"] = int64(results.Page)
			ctx["pages"] = int64(results.Pages)
			return ctx, nil
		}, w, r)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			Output404(w, r, err.Error())
			return
		}
		Output500(w, r, err)
		return
	}
	if middlewares.HeadersSent(w) {
		return
	}
	render.HTML(w, r, string(parsed))
})
