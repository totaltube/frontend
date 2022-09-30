package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var Category = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*site.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	var pageStr = chi.URLParam(r, "page")
	if pageStr == "" {
		pageStr = r.URL.Query().Get(config.Params.Page)
	}
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.ParseInt(pageStr, 10, 16)
	if page <= 0 {
		page = 1
	}
	categorySlug := helpers.FirstNotEmpty(chi.URLParam(r, "slug"), r.URL.Query().Get(config.Params.CategorySlug))
	categoryId, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "id"), r.URL.Query().Get(config.Params.CategoryId)), 10, 64)
	if categoryId == 0 && categorySlug == "" {
		Output404(w, r, "category not found")
		return
	}
	sortBy := r.URL.Query().Get(config.Params.SortBy)
	if sortBy == config.Params.SortByDate {
		sortBy = "dated"
	} else if sortBy == config.Params.SortByDuration {
		sortBy = "duration"
	} else if sortBy == config.Params.SortByViews {
		sortBy = "views"
	} else if sortBy == config.Params.SortByRand {
		sortBy = "rand"
	} else {
		sortBy = ""
	}
	sortByViewsTimeframe := r.URL.Query().Get(config.Params.SortByViewsTimeframe)
	channelId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ChannelId), 10, 64)
	channelSlug := r.URL.Query().Get(config.Params.ChannelSlug)
	modelId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ModelId), 10, 64)
	modelSlug := r.URL.Query().Get(config.Params.ModelSlug)
	durationFrom, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationGte), 10, 64)
	durationTo, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationLt), 10, 64)
	customContext := generateCustomContext(w, r, "category")
	cacheKey := "category:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d",
			hostName, langId, categoryId, categorySlug, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo),
	)
	filtered := channelId > 0 || channelSlug != "" || modelId > 0 || modelSlug != "" || sortBy != "" ||
		durationTo > 0 || durationFrom > 0
	cacheTtl := time.Second * 5
	if page > 1 || filtered {
		cacheTtl = time.Minute * 5
	}
	ip := r.Context().Value("ip").(string)
	userAgent := r.Header.Get("User-Agent")
	parsed, err := site.ParseTemplate("category", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			categoryInfoCacheKey := fmt.Sprintf("in:cinfo:%d:%s:%s", categoryId, categorySlug, langId)
			categoryInfoCacheTtl := time.Hour*24 + time.Duration(rand.Intn(3600*6))*time.Second
			categoryInfoCached, err := db.GetCachedTimeout(categoryInfoCacheKey, categoryInfoCacheTtl, time.Hour*4, func() ([]byte, error) {
				_, rawResponse, err := api.CategoryInfo(hostName, langId, categoryId, categorySlug)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			categoryInfo := new(types.CategoryResult)
			err = json.Unmarshal(categoryInfoCached, categoryInfo)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			var results *types.ContentResults
			if filtered {
				results, _, err = api.Content(hostName, api.ContentParams{
					Ip:           net.ParseIP(ip),
					Lang:         langId,
					Page:         page,
					CategoryId:   categoryId,
					CategorySlug: categorySlug,
					ChannelId:    channelId,
					ChannelSlug:  channelSlug,
					ModelId:      modelId,
					ModelSlug:    modelSlug,
					Sort:         api.SortBy(sortBy),
					Timeframe:    sortByViewsTimeframe,
					DurationGte:  durationFrom,
					DurationLt:   durationTo,
					UserAgent:    userAgent,
				})
			} else {
				ctx["count"] = true
				results, err = api.Category(hostName, langId, categoryId, categorySlug, page)
			}
			if err != nil {
				return ctx, err
			}
			ctx["category"] = categoryInfo
			ctx["content"] = results
			ctx["total"] = results.Total
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
	render.HTML(w, r, string(parsed))
})

func getCategoryTopFunc(hostName string, langId string) func(args ...interface{}) *types.ContentResults {
	return func(args ...interface{}) *types.ContentResults {
		parsingName := true
		var categoryId int64
		var categorySlug string
		var page int64
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
				page, _ = strconv.ParseInt(val, 10, 16)
			case "category_id", "id":
				categoryId, _ = strconv.ParseInt(val, 10, 32)
			case "category_slug", "slug":
				categorySlug = val
			}
		}
		if categoryId == 0 && categorySlug == "" {
			log.Println("error getting top category content - need to set category_id or category_slug param")
			return nil
		}
		if results, err := api.Category(hostName, langId, categoryId, categorySlug, page); err != nil {
			log.Println("error getting category top content: ", err)
			return nil
		} else {
			return results
		}
	}
}
