package handlers

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var Search = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*site.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	searchQuery, _ := url.PathUnescape(strings.ReplaceAll(chi.URLParam(r, "query"), "+", "%20"))
	if searchQuery == "" {
		searchQuery = r.URL.Query().Get(config.Params.SearchQuery)
	}
	if searchQuery == "" {
		panic(errors.New("search query not set"))
	}
	isNatural, _ := strconv.ParseBool(config.Params.SearchNatural)
	modelId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ModelId), 10, 64)
	modelSlug := r.URL.Query().Get(config.Params.ModelSlug)
	categorySlug := r.URL.Query().Get(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.CategoryId), 10, 64)
	sortBy := r.URL.Query().Get(config.Params.SortBy)
	sortByTimeframe := r.URL.Query().Get(config.Params.SortByViewsTimeframe)
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
	customContext := generateCustomContext(w, r, "search")
	cacheKey := "search:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%d:%d:%s:%d:%d:%d:%s:%s:%s",
			hostName, langId, page, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug, sortBy, searchQuery),
	)
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	userAgent := r.Header.Get("User-Agent")
	cacheTtl := time.Minute * 15
	parsed, err := site.ParseTemplate("search", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			var results *types.ContentResults
			var err error
			results, _, err = api.Content(hostName, api.ContentParams{
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
			})
			if err != nil {
				return ctx, err
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
