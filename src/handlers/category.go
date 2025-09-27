package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
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

var Category = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value(types.ContextKeyLang).(string)
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
	rotationParamsList := strings.Split(r.URL.Query().Get(config.Params.Rotation), "-")
	var rotationParams = rotationParams{
		Type:       types.CountTypeNone,
		ContentId:  0,
		CategoryId: 0,
		ThumbId:    -1,
		Position:   -1,
	}
	useTrade := false
	for _, param := range rotationParamsList {
		param_parts := strings.Split(param, ":")
		if len(param_parts) == 2 {
			switch param_parts[0] {
			case config.Params.CountType:
				switch param_parts[1] {
				case config.Params.CountTypeCategory:
					rotationParams.Type = types.CountTypeCategory
				case config.Params.CountTypeTopCategories:
					rotationParams.Type = types.CountTypeTopCategories
				case config.Params.CountTypeTopContent:
					rotationParams.Type = types.CountTypeTopContent
				}
			case config.Params.ContentId:
				rotationParams.ContentId, _ = strconv.ParseInt(param_parts[1], 10, 64)
			case config.Params.CategoryId:
				rotationParams.CategoryId, _ = strconv.ParseInt(param_parts[1], 10, 64)
			case config.Params.CountThumbId:
				rotationParams.ThumbId, _ = strconv.ParseInt(param_parts[1], 10, 64)
			case config.Params.CountPosition:
				rotationParams.Position, _ = strconv.ParseInt(param_parts[1], 10, 64)
			case config.Params.RotationTrade:
				useTrade = true
			case config.Params.Skim:
				rotationParams.Skim = param_parts[1]
			}
		}
	}
	if (rotationParams.Type != types.CountTypeNone && page == 1) || useTrade {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		toReturn := handleRotation(rotationParams, useTrade, config, r, w)
		if toReturn {
			return
		}
	}
	categorySlug := helpers.FirstNotEmpty(chi.URLParam(r, "slug"), r.URL.Query().Get(config.Params.CategorySlug))
	categoryId, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "id"), r.URL.Query().Get(config.Params.CategoryId)), 10, 64)
	if categoryId == 0 && categorySlug == "" {
		Output404(w, r, "category not found")
		return
	}
	if categoryId > 0 && config.Routes.IdXorKey > 0 {
		categoryId = categoryId ^ config.Routes.IdXorKey
	}
	sortBy := r.URL.Query().Get(config.Params.SortBy)
	switch sortBy {
	case config.Params.SortByDate:
		sortBy = "dated"
	case config.Params.SortByDuration:
		sortBy = "duration"
	case config.Params.SortByViews:
		sortBy = "views"
	case config.Params.SortByRand:
		sortBy = "rand"
	default:
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
	ip := net.ParseIP(r.Context().Value(types.ContextKeyIp).(string))
	groupId := internal.DetectCountryGroup(ip).Id
	amount := config.General.CategoryResultsPerPage
	cacheKey := "category:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d:%d:%d",
			hostName, langId, categoryId, categorySlug, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, amount, groupId),
	)
	filtered := channelId > 0 || channelSlug != "" || modelId > 0 || modelSlug != "" || sortBy != "" ||
		durationTo > 0 || durationFrom > 0
	cacheTtl := time.Minute * 3
	if page > 1 || filtered {
		cacheTtl = time.Minute * 30
	}

	userAgent := r.Header.Get("User-Agent")
	pageTtl := 0 * time.Second
	randomizeRatio := config.General.RandomizeRatio
	if randomizeRatio < 0 {
		randomizeRatio = internal.Config.General.RandomizeRatio
	}
	if randomizeRatio <= 0 {
		pageTtl = time.Minute * 3
	}
	parsed, err := site.ParseTemplate("category", path, config, customContext, nocache, cacheKey, pageTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			categoryInfoCacheKey := fmt.Sprintf("in:cinfo:%d:%s:%s", categoryId, categorySlug, langId)
			categoryInfoCacheTtl := time.Hour*24 + time.Duration(rand.Intn(3600*6))*time.Second
			categoryInfoCached, err := db.GetCachedTimeout(categoryInfoCacheKey, categoryInfoCacheTtl, time.Hour*4, func() ([]byte, error) {
				_, rawResponse, err := api.CategoryInfo(hostName, langId, categoryId, categorySlug)
				return rawResponse, err
			}, nocache)
			if err != nil {
				if !strings.Contains(err.Error(), "favicon.ico") {
					log.Println(err, config.Hostname, ip)
				}
				return ctx, err
			}
			categoryInfo := new(types.CategoryResult)
			err = json.Unmarshal(categoryInfoCached, categoryInfo)
			if err != nil {
				log.Println(err, config.Hostname)
				return ctx, err
			}
			var results = new(types.ContentResults)
			if filtered {
				var response []byte
				response, err = db.GetCachedTimeout(cacheKey+":data", cacheTtl, cacheTtl, func() ([]byte, error) {
					return api.ContentRaw(hostName, api.ContentParams{
						Lang:         langId,
						Page:         page,
						Ip:           ip,
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
						Amount:       amount,
					})
				}, nocache)
				if err != nil {
					return ctx, err
				}
				err = json.Unmarshal(response, results)
			} else {
				ctx["count"] = true
				var response []byte
				response, err = db.GetCachedTimeout(cacheKey+":data", cacheTtl, cacheTtl, func() ([]byte, error) {
					return api.CategoryRaw(hostName, langId, categoryId, categorySlug, page, groupId)
				}, nocache)
				if err != nil {
					return ctx, err
				}
				err = json.Unmarshal(response, results)
			}
			if err != nil {
				return ctx, err
			}
			if len(results.Items) == 0 && page > 1 {
				return ctx, fmt.Errorf("not found")
			}
			if page == 1 && randomizeRatio > 0 {
				helpers.RandomizeItems(results.Items, randomizeRatio)
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
			if strings.Contains(err.Error(), "not found") {
				if categorySlug != "" && config.General.DeletedTaxonomiesToSearch || internal.Config.General.DeletedTaxonomiesToSearch {
					redirectType := 302
					if config.General.DeletedTaxonomiesToSearchPermanent || internal.Config.General.DeletedTaxonomiesToSearchPermanent {
						redirectType = 301
					}
					link := site.GetLink("search", config, hostName, langId, false, "search_query", strings.ReplaceAll(categorySlug, "-", "+"))
					http.Redirect(w, r, link, redirectType)
					if internal.Config.General.EnableAccessLog {
						log.Printf("Redirected to search: %s", link)
					}
					return
				}
				Output404(w, r, err.Error())
				return
			}
		}
		Output500(w, r, err)
		return
	}
	if middlewares.HeadersSent(w) {
		return
	}
	render.HTML(w, r, string(parsed))
})

func getCategoryFunc(hostName string, langId string) func(args ...interface{}) *types.CategoryResult {
	return func(args ...interface{}) *types.CategoryResult {
		parsingName := true
		var categoryId int64
		var categorySlug string
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
			case "category_id", "id":
				categoryId, _ = strconv.ParseInt(val, 10, 32)
			case "category_slug", "slug":
				categorySlug = val
			}
		}
		if categoryId == 0 && categorySlug == "" {
			log.Println("error getting category content - need to set category_id or category_slug param")
			return nil
		}
		if results, _, err := api.CategoryInfo(hostName, langId, categoryId, categorySlug); err != nil {
			log.Println("error getting category content: ", err)
			return nil
		} else {
			return results
		}
	}
}
func getCategoryTopFunc(hostName string, langId string, groupId int64, config *types.Config) func(args ...interface{}) *types.ContentResults {
	return func(args ...interface{}) *types.ContentResults {
		parsingName := true
		var categoryId int64
		var categorySlug string
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
				page, _ = strconv.ParseInt(val, 10, 16)
			case "category_id", "id":
				categoryId, _ = strconv.ParseInt(val, 10, 32)
			case "category_slug", "slug":
				categorySlug = val
			case "group_id":
				groupId, _ = strconv.ParseInt(val, 10, 32)
			}
		}
		if categoryId == 0 && categorySlug == "" {
			log.Println("error getting top category content - need to set category_id or category_slug param")
			return nil
		}
		if results, err := api.Category(hostName, langId, categoryId, categorySlug, page, groupId); err != nil {
			log.Println("error getting category top content: ", err)
			return nil
		} else {
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
}
