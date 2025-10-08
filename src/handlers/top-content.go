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

var TopContent = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	ip := r.Context().Value(types.ContextKeyIp).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value(types.ContextKeyLang).(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(w, r, "top-content")
	var groupId = internal.DetectCountryGroup(net.ParseIP(ip)).Id
	cacheKey := fmt.Sprintf("top-content:%s:%s:%d:%d", hostName, langId, page, groupId)
	var cacheTtl types.Duration
	if config.CacheTimeouts.TopContent != nil {
		cacheTtl = *config.CacheTimeouts.TopContent
	} else {
		cacheTtl = internal.Config.CacheTimeouts.TopContent
	}
	if page > 1 {
		if config.CacheTimeouts.TopContentPagination != nil {
			cacheTtl = *config.CacheTimeouts.TopContentPagination
		} else {
			cacheTtl = internal.Config.CacheTimeouts.TopContentPagination
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
	parsed, err := site.ParseTemplate("top-content", path, config, customContext, nocache, cacheKey, pageTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			var results = new(types.ContentResults)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", time.Duration(cacheTtl), time.Duration(cacheTtl), func() ([]byte, error) {
				bt, err := api.TopContentRaw(hostName, langId, page, groupId)
				return bt, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			err = json.Unmarshal(response, results)
			if err != nil {
				return ctx, err
			}
			if len(results.Items) == 0 && page > 1 {
				return ctx, fmt.Errorf("not found")
			}
			if page == 1 {
				ctx["count"] = true
			}
			if page == 1 && randomizeRatio > 0 {
				helpers.RandomizeItems(results.Items, randomizeRatio)
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

func getTopContentFunc(hostName string, langId string, groupId int64, config *types.Config) func(args ...interface{}) *types.ContentResults {
	return func(args ...interface{}) *types.ContentResults {
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
		results, err := api.TopContent(hostName, langId, page, groupId)
		if err != nil {
			log.Println("can't get top content:", err)
			return nil
		}
		randomizeRatio := config.General.RandomizeRatio
		if randomizeRatio < 0 {
			randomizeRatio = internal.Config.General.RandomizeRatio
		}
		if randomizeRatio > 0 {
			helpers.RandomizeItems(results.Items, randomizeRatio)
		}
		return results
	}
}
