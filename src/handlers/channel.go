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

var Channel = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*types.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	modelId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ModelId), 10, 64)
	if modelId > 0 && config.Routes.IdXorKey > 0 {
		modelId = modelId ^ config.Routes.IdXorKey
	}
	modelSlug := r.URL.Query().Get(config.Params.ModelSlug)
	categorySlug := r.URL.Query().Get(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.CategoryId), 10, 64)
	if categoryId > 0 && config.Routes.IdXorKey > 0 {
		categoryId = categoryId ^ config.Routes.IdXorKey
	}
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
	channelId, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "id"), r.URL.Query().Get(config.Params.ChannelId)), 10, 64)
	channelSlug := helpers.FirstNotEmpty(chi.URLParam(r, "slug"), r.URL.Query().Get(config.Params.ChannelSlug))
	if channelId == 0 && channelSlug == "" {
		Output404(w, r, "channel not found")
		return
	}
	if channelId > 0 && config.Routes.IdXorKey > 0 {
		channelId = channelId ^ config.Routes.IdXorKey
	}
	durationGte, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationGte), 10, 64)
	durationLt, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationLt), 10, 64)
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "channel")
	amount := config.General.ChannelResultsPerPage
	if amount == 0 {
		amount = config.General.DefaultResultsPerPage
	}
	cacheKey := "channel:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d:%d:%s:%d:%d",
			hostName, langId, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationGte, durationLt, categoryId, categorySlug, groupId, amount),
	)
	cacheTtl := time.Minute * 15
	userAgent := r.Header.Get("User-Agent")
	parsed, err := site.ParseTemplate("channel", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			channelInfoCacheKey := fmt.Sprintf("in:chinfo:%d:%s:%s", channelId, channelSlug, langId)
			channelInfoCacheTtl := time.Hour*24 + time.Duration(rand.Intn(3600*6))*time.Second
			channelInfoCached, err := db.GetCachedTimeout(channelInfoCacheKey, channelInfoCacheTtl, time.Hour*4, func() ([]byte, error) {
				_, rawResponse, err := api.ChannelInfo(hostName, langId, channelId, channelSlug)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			channelInfo := new(types.ChannelResult)
			err = json.Unmarshal(channelInfoCached, channelInfo)
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
					DurationGte:  durationGte,
					DurationLt:   durationLt,
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
			ctx["channel"] = channelInfo
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
			if channelSlug != "" && config.General.DeletedTaxonomiesToSearch || internal.Config.General.DeletedTaxonomiesToSearch {
				redirectType := 302
				if config.General.DeletedTaxonomiesToSearchPermanent || internal.Config.General.DeletedTaxonomiesToSearchPermanent {
					redirectType = 301
				}
				link := site.GetLink("search", config, hostName, langId, false, "search_query", strings.ReplaceAll(channelSlug, "-", "+"))
				http.Redirect(w, r, link, redirectType)
				if internal.Config.General.EnableAccessLog {
					log.Println("Redirected to search page for deleted channel:", channelSlug)
				}
				return
			}
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
