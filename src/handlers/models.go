package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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

var Models = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value(types.ContextKeyLang).(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	// can be title, total, popular
	sortBy := helpers.FirstNotEmpty(chi.URLParam(r, "sort"), r.URL.Query().Get(config.Params.SortBy), "title")
	query := r.URL.Query().Get(config.Params.SearchQuery)
	amount := config.General.ModelsPerPage
	ip := r.Context().Value(types.ContextKeyIp).(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "models")
	cacheKey := "models:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%s:%d:%d",
			hostName, langId, page, sortBy, query, amount, groupId),
	)
	var cacheTtl types.Duration
	if config.CacheTimeouts.Models != nil {
		cacheTtl = *config.CacheTimeouts.Models
	} else {
		cacheTtl = internal.Config.CacheTimeouts.Models
	}
	if page > 1 {
		if config.CacheTimeouts.ModelsPagination != nil {
			cacheTtl = *config.CacheTimeouts.ModelsPagination
		} else {
			cacheTtl = internal.Config.CacheTimeouts.ModelsPagination
		}
	}
	parsed, err := site.ParseTemplate("models", path, config, customContext, nocache, cacheKey, time.Duration(cacheTtl),
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			var results = new(types.ModelResults)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", time.Duration(cacheTtl), time.Duration(cacheTtl), func() ([]byte, error) {
				return api.ModelsListRaw(hostName, langId, page, api.SortBy(sortBy), int64(amount), query, groupId)
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
	if middlewares.HeadersSent(w) {
		return
	}
	render.HTML(w, r, string(parsed))
})

func getModelFunc(hostName string, langId string, groupId int64) func(args ...interface{}) *types.ModelResult {
	return func(args ...interface{}) *types.ModelResult {
		parsingName := true
		var modelId int64
		var modelSlug string
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
			case "id", "model_id":
				modelId, _ = strconv.ParseInt(val, 10, 64)
			case "slug", "model_slug":
				modelSlug = val
			}
		}
		results, _, err := api.ModelInfo(hostName, langId, modelId, modelSlug, groupId)
		if err != nil {
			log.Println("can't get model info:", err)
			return nil
		}
		return results
	}
}
func getModelsListFunc(hostName string, langId string, defaultAmount int64, groupId int64) func(args ...interface{}) *types.ModelResults {
	return func(args ...interface{}) *types.ModelResults {
		parsingName := true
		var amount = defaultAmount
		var page int64
		var sortBy = api.SortTitle
		var searchQuery = ""
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
				page, _ = strconv.ParseInt(val, 10, 64)
			case "sort":
				sortBy = api.SortBy(val)
			case "amount":
				amount, _ = strconv.ParseInt(val, 10, 64)
			case "search_query":
				searchQuery = val
			}
		}
		results, _, err := api.ModelsList(hostName, langId, page, sortBy, amount, searchQuery, groupId)
		if err != nil {
			log.Println("can't get models list:", err)
			return nil
		}
		return results
	}
}
