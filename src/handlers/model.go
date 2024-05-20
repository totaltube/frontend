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
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var Model = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*types.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	modelId, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "id"), r.URL.Query().Get(config.Params.ModelId)), 10, 64)
	modelSlug := helpers.FirstNotEmpty(chi.URLParam(r, "slug"), r.URL.Query().Get(config.Params.ModelSlug))
	if modelId == 0 && modelSlug == "" {
		Output404(w, r, "model not found")
		return
	}
	categorySlug := r.URL.Query().Get(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.CategoryId), 10, 64)
	sortBy := helpers.FirstNotEmpty(r.URL.Query().Get(config.Params.SortBy), "dated")
	sortByViewsTimeframe := r.URL.Query().Get(config.Params.SortByViewsTimeframe)
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
	channelId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ChannelId), 10, 64)
	channelSlug := r.URL.Query().Get(config.Params.ChannelSlug)
	durationFrom, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationGte), 10, 64)
	durationTo, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationLt), 10, 64)
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "model")
	amount := config.General.ModelResultsPerPage
	if amount == 0 {
		amount = config.General.DefaultResultsPerPage
	}
	cacheKey := "model:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d:%d:%s:%d:%d",
			hostName, langId, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug, groupId, amount),
	)
	userAgent := r.Header.Get("User-Agent")
	cacheTtl := time.Minute * 15

	parsed, err := site.ParseTemplate("model", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			modelInfoCacheKey := fmt.Sprintf("in:minfo:%d:%s:%s", modelId, modelSlug, langId)
			modelInfoCacheTtl := time.Hour*24 + time.Duration(rand.Intn(3600*6))*time.Second
			modelInfoCached, err := db.GetCachedTimeout(modelInfoCacheKey, modelInfoCacheTtl, time.Hour*4, func() ([]byte, error) {
				_, rawResponse, err := api.ModelInfo(hostName, langId, modelId, modelSlug, groupId)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			modelInfo := new(types.ModelResult)
			err = json.Unmarshal(modelInfoCached, modelInfo)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			var results = new(types.ContentResults)
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", cacheTtl, cacheTtl, func() ([]byte, error) {
				return api.ContentRaw(hostName, api.ContentParams{
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
					GroupId:      groupId,
					Amount:       amount,
					Ip:           net.ParseIP(ip),
				})
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
			ctx["model"] = modelInfo
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
				if modelSlug != "" && config.General.DeletedTaxonomiesToSearch || internal.Config.General.DeletedTaxonomiesToSearch {
					redirectType := 302
					if config.General.DeletedTaxonomiesToSearchPermanent || internal.Config.General.DeletedTaxonomiesToSearchPermanent {
						redirectType = 301
					}
					link := site.GetLink("search", config, hostName, langId, false, "search_query", strings.ReplaceAll(modelSlug, "-", "+"))
					http.Redirect(w, r, link, redirectType)
					return
				}
				Output404(w, r, err.Error())
				return
			}
		}
		Output500(w, r, err)
		return
	}
	render.HTML(w, r, string(parsed))
})
