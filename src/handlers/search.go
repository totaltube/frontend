package handlers

import (
	"encoding/base64"
	"fmt"
	"html"
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
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var Search = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))

	langId := r.Context().Value(types.ContextKeyLang).(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	// Try path parameters first
	searchQuery, _ := url.PathUnescape(strings.ReplaceAll(chi.URLParam(r, "query"), "+", "%20"))
	if searchQuery == "" {
		// Try query_base64 from path
		if queryBase64Path := chi.URLParam(r, "query_base64"); queryBase64Path != "" {
			if decoded, err := base64.RawURLEncoding.DecodeString(queryBase64Path); err == nil {
				searchQuery = string(decoded)
			}
		}
	}
	if searchQuery == "" {
		// Try query_htmlentities from path
		if queryHtmlEntitiesPath := chi.URLParam(r, "query_htmlentities"); queryHtmlEntitiesPath != "" {
			if unescaped, err := url.PathUnescape(queryHtmlEntitiesPath); err == nil {
				searchQuery = html.UnescapeString(unescaped)
			} else {
				searchQuery = html.UnescapeString(queryHtmlEntitiesPath)
			}
		}
	}
	// If not found in path, try query parameters
	if searchQuery == "" {
		searchQuery = r.URL.Query().Get(config.Params.SearchQuery)
	}
	searchQuery = strings.TrimSpace(strings.ReplaceAll(searchQuery, "  ", " "))
	if searchQuery == "" {
		Output404(w, r, "search query not set")
		return
	}
	isNatural, _ := strconv.ParseBool(config.Params.SearchNatural)
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
	sortBy := r.URL.Query().Get(config.Params.SortBy)
	sortByTimeframe := r.URL.Query().Get(config.Params.SortByViewsTimeframe)
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
	channelId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ChannelId), 10, 64)
	if channelId > 0 && config.Routes.IdXorKey > 0 {
		channelId = channelId ^ config.Routes.IdXorKey
	}
	channelSlug := r.URL.Query().Get(config.Params.ChannelSlug)
	durationFrom, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationGte), 10, 64)
	durationTo, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.DurationLt), 10, 64)
	customContext := generateCustomContext(w, r, "search")
	amount := config.General.SearchResultsPerPage
	if amount == 0 {
		amount = config.General.DefaultResultsPerPage
	}
	cacheKey := fmt.Sprintf("%s:%s:%d:%s:%d:%d:%s:%d:%d:%d:%s:%s:%s:%d",
		hostName, langId, page, channelSlug, channelId,
		modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug, sortBy, searchQuery, amount)
	for _, param := range config.General.CacheKeyQueryParams {
		v := r.URL.Query().Get(param)
		if v != "" {
			cacheKey += fmt.Sprintf(":%s:%s", param, v)
		}
	}
	cacheKey = "search:" + helpers.Md5Hash(cacheKey)
	ip := r.Context().Value(types.ContextKeyIp).(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	userAgent := r.Header.Get("User-Agent")
	var cacheTtl types.Duration
	if config.CacheTimeouts.Search != nil {
		cacheTtl = *config.CacheTimeouts.Search
	} else {
		cacheTtl = internal.Config.CacheTimeouts.Search
	}
	if page > 1 {
		if config.CacheTimeouts.SearchPagination != nil {
			cacheTtl = *config.CacheTimeouts.SearchPagination
		} else {
			cacheTtl = internal.Config.CacheTimeouts.SearchPagination
		}
	}
	started := time.Now()
	parsed, err := site.ParseTemplate("search", path, config, customContext, nocache, cacheKey, time.Duration(cacheTtl),
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			var results *types.ContentResults
			var err error
			results, _, err = api.Content(config, api.ContentParams{
				Ip:           net.ParseIP(ip),
				SearchQuery:  searchQuery,
				IsNatural:    isNatural,
				Lang:         langId,
				Page:         page,
				CategoryId:   categoryId,
				CategorySlug: categorySlug,
				ChannelId:    channelId,
				ChannelSlug:  channelSlug,
				ModelId:      modelId,
				ModelSlug:    modelSlug,
				Sort:         api.SortBy(sortBy),
				Timeframe:    sortByTimeframe,
				DurationGte:  durationFrom,
				DurationLt:   durationTo,
				UserAgent:    userAgent,
				GroupId:      groupId,
				Amount:       amount,
			})
			if err != nil {
				return ctx, err
			}
			if len(results.Items) == 0 && page > 1 {
				return ctx, fmt.Errorf("not found")
			}
			ctx["search_query"] = searchQuery
			ctx["content"] = results
			ctx["total"] = results.Total
			ctx["from"] = int64(results.From)
			ctx["to"] = int64(results.To)
			ctx["page"] = int64(results.Page)
			ctx["pages"] = int64(results.Pages)
			return ctx, nil
		}, w, r)
	elapsed := time.Since(started)
	if elapsed > time.Second*5 {
		log.Printf("Long processing search query \"%s\" of site %s: %v, user Agent: %s", searchQuery, hostName, elapsed, r.Header.Get("User-Agent"))
	}
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

func getTopSearchesFunc(config *types.Config, langId string) func(args ...any) []types.TopSearch {
	return func(args ...any) []types.TopSearch {
		currentName := ""
		parsingName := true
		amount := int64(10)
		for _, arg := range args {
			if parsingName {
				currentName = fmt.Sprintf("%v", arg)
				parsingName = false
			} else {
				v := fmt.Sprintf("%v", arg)
				if currentName == "amount" {
					amount, _ = strconv.ParseInt(v, 10, 64)
				}
			}
		}
		results, _, err := api.TopSearches(config, langId, int64(amount))
		if err != nil {
			log.Println(err)
			return nil
		}
		return results
	}
}

func getRandomSearchesFunc(config *types.Config, langId string) func(args ...any) []types.TopSearch {
	return func(args ...any) []types.TopSearch {
		currentName := ""
		parsingName := true
		amount := int64(10)
		minSearches := int64(0)
		for _, arg := range args {
			if parsingName {
				currentName = fmt.Sprintf("%v", arg)
				parsingName = false
			} else {
				v := fmt.Sprintf("%v", arg)
				switch currentName {
				case "amount":
					amount, _ = strconv.ParseInt(v, 10, 64)
				case "min_searches":
					minSearches, _ = strconv.ParseInt(v, 10, 64)
				}
			}
		}
		results, _, err := api.RandomSearches(config, langId, int64(amount), int64(minSearches))
		if err != nil {
			log.Println(err)
			return nil
		}
		return results
	}
}
