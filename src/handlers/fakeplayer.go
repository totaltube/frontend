package handlers

import (
	"encoding/json"
	"errors"
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

var FakePlayer = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value(types.ContextKeyLang).(string)
	slug := chi.URLParam(r, "slug")
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id == 0 && slug == "" {
		Output404(w, r, "content item not found")
		return
	}
	if id > 0 && config.Routes.IdXorKey > 0 {
		id = id ^ config.Routes.IdXorKey
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	customContext := generateCustomContext(w, r, "fake-player")
	params := customContext["params"].(map[string]string)
	ip := r.Context().Value(types.ContextKeyIp).(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	relatedRandomizeLast := 0
	if config.Related.Randomize != nil {
		relatedRandomizeLast = *config.Related.Randomize
	} else if internal.Config.Related.Randomize != nil {
		relatedRandomizeLast = *internal.Config.Related.Randomize
	}
	relatedParams := &api.RelatedParams{
		TitleTranslated:              config.Related.TitleTranslated,
		TitleTranslatedMinTermFreq:   config.Related.TitleTranslatedMinTermFreq,
		TitleTranslatedMaxQueryTerms: config.Related.TitleTranslatedMaxQueryTerms,
		TitleTranslatedBoost:         config.Related.TitleTranslatedBoost,
		Title:                        config.Related.Title,
		TitleMinTermFreq:             config.Related.TitleMinTermFreq,
		TitleMaxQueryTerms:           config.Related.TitleMaxQueryTerms,
		TitleBoost:                   config.Related.TitleBoost,
		Tags:                         config.Related.Tags,
		TagsMinTermFreq:              config.Related.TagsMinTermFreq,
		TagsMaxQueryTerms:            config.Related.TagsMaxQueryTerms,
		TagsBoost:                    config.Related.TagsBoost,
		RandomizeLast:                relatedRandomizeLast,
	}
	cacheKey := "fake-player:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%v:%d:%d", hostName, langId, id, slug, orfl, relatedAmount, groupId),
	)

	var cacheTtl types.Duration
	if config.CacheTimeouts.ContentItem != nil {
		cacheTtl = *config.CacheTimeouts.ContentItem
	} else {
		cacheTtl = internal.Config.CacheTimeouts.ContentItem
	}
	parsed, err := site.ParseTemplate("fake-player", path, config, customContext, nocache, cacheKey, time.Duration(cacheTtl),
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			var results = new(types.ContentItemResult)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", time.Duration(cacheTtl), time.Duration(cacheTtl), func() ([]byte, error) {
				return api.ContentItemRaw(hostName, langId, slug, id, orfl, int64(relatedAmount), groupId, relatedParams)
			}, nocache)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return ctx, errors.New("content item not found")
				}
				return ctx, err
			}
			err = json.Unmarshal(response, results)
			if err != nil {
				return ctx, err
			}
			if slug != "" && results.Slug != slug {
				// Rephrased title and slug, need to make a 301 redirect
				categorySlug := "default"
				if len(results.Categories) > 0 {
					categorySlug = results.Categories[0].Slug
				}
				var args = make([]interface{}, 0, 10)
				args = append(args,
					"slug", results.Slug,
					"id", results.Id,
					"category", categorySlug,
				)
				for k := range r.URL.Query() {
					args = append(args, k, r.URL.Query().Get(k))
				}
				return nil, redirectErr{
					url:  site.GetLink("content", config, hostName, langId, false, args...),
					code: 301,
				}
			}
			ctx["content_item"] = results
			ctx["related"] = results.Related
			// let's add first category name to params
			if len(results.Categories) > 0 {
				params["category"] = results.Categories[0].Slug
			} else {
				params["category"] = "default"
			}
			id := results.Id
			if config.Routes.IdXorKey > 0 {
				id = id ^ config.Routes.IdXorKey
			}
			params["id"] = strconv.FormatInt(id, 10)
			params["slug"] = results.Slug
			ctx["params"] = params
			return ctx, nil
		}, w, r)
	if err != nil {
		if rErr, ok := err.(redirectErr); ok {
			http.Redirect(w, r, rErr.url, rErr.code)
			if internal.Config.General.EnableAccessLog {
				log.Printf("Redirected to %s", rErr.url)
			}
			return
		}
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
	//w.Header().Set("X-Robots-Tag", "noindex")
	render.HTML(w, r, string(parsed))
})
