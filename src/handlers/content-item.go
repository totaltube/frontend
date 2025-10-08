package handlers

import (
	"encoding/json"
	"errors"
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

type redirectErr struct {
	code int
	url  string
}

func (r redirectErr) Error() string {
	return fmt.Sprintf("%d redirect to %s", r.code, r.url)
}

type rotationParams struct {
	Type       types.CountType
	ContentId  int64
	CategoryId int64
	ThumbId    int64
	Position   int64
	Skim       string
}

var ContentItem = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value(types.ContextKeyLang).(string)
	slug, _ := url.PathUnescape(chi.URLParam(r, "slug"))
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
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
	if rotationParams.Type != types.CountTypeNone || useTrade {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		toReturn := handleRotation(rotationParams, useTrade, config, r, w)
		if toReturn {
			return
		}
	}
	if id == 0 && slug == "" {
		Output404(w, r, "content item not found")
		return
	}
	if id > 0 && config.Routes.IdXorKey > 0 {
		id = id ^ config.Routes.IdXorKey
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	ip := r.Context().Value(types.ContextKeyIp).(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "content-item")
	params := customContext["params"].(map[string]string)
	relatedTitleTranslated := config.Related.TitleTranslated
	if relatedTitleTranslated == nil {
		relatedTitleTranslated = internal.Config.Related.TitleTranslated
	}
	relatedTitleTranslatedMinTermFreq := config.Related.TitleTranslatedMinTermFreq
	if relatedTitleTranslatedMinTermFreq == nil {
		relatedTitleTranslatedMinTermFreq = internal.Config.Related.TitleTranslatedMinTermFreq
	}
	relatedTitleTranslatedMaxQueryTerms := config.Related.TitleTranslatedMaxQueryTerms
	if relatedTitleTranslatedMaxQueryTerms == nil {
		relatedTitleTranslatedMaxQueryTerms = internal.Config.Related.TitleTranslatedMaxQueryTerms
	}
	relatedTitleTranslatedBoost := config.Related.TitleTranslatedBoost
	if relatedTitleTranslatedBoost == nil {
		relatedTitleTranslatedBoost = internal.Config.Related.TitleTranslatedBoost
	}
	relatedTitle := config.Related.Title
	if relatedTitle == nil {
		relatedTitle = internal.Config.Related.Title
	}
	relatedTitleMinTermFreq := config.Related.TitleMinTermFreq
	if relatedTitleMinTermFreq == nil {
		relatedTitleMinTermFreq = internal.Config.Related.TitleMinTermFreq
	}
	relatedTitleMaxQueryTerms := config.Related.TitleMaxQueryTerms
	if relatedTitleMaxQueryTerms == nil {
		relatedTitleMaxQueryTerms = internal.Config.Related.TitleMaxQueryTerms
	}
	relatedTitleBoost := config.Related.TitleBoost
	if relatedTitleBoost == nil {
		relatedTitleBoost = internal.Config.Related.TitleBoost
	}
	relatedTags := config.Related.Tags
	if relatedTags == nil {
		relatedTags = internal.Config.Related.Tags
	}
	relatedTagsMinTermFreq := config.Related.TagsMinTermFreq
	if relatedTagsMinTermFreq == nil {
		relatedTagsMinTermFreq = internal.Config.Related.TagsMinTermFreq
	}
	relatedTagsMaxQueryTerms := config.Related.TagsMaxQueryTerms
	if relatedTagsMaxQueryTerms == nil {
		relatedTagsMaxQueryTerms = internal.Config.Related.TagsMaxQueryTerms
	}
	relatedTagsBoost := config.Related.TagsBoost
	if relatedTagsBoost == nil {
		relatedTagsBoost = internal.Config.Related.TagsBoost
	}
	cacheKey := fmt.Sprintf("%s:%s:%d:%s:%v:%d:%d:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v", hostName, langId, id, slug, orfl, relatedAmount, groupId,
		relatedTitleTranslated, relatedTitleTranslatedMinTermFreq, relatedTitleTranslatedMaxQueryTerms, relatedTitleTranslatedBoost,
		relatedTitle, relatedTitleMinTermFreq, relatedTitleMaxQueryTerms, relatedTitleBoost,
		relatedTags, relatedTagsMinTermFreq, relatedTagsMaxQueryTerms, relatedTagsBoost,
	)
	for _, param := range config.General.CacheKeyQueryParams {
		v := r.URL.Query().Get(param)
		if v != "" {
			cacheKey += fmt.Sprintf(":%s:%s", param, v)
		}
	}
	cacheKey = "content-item:" + helpers.Md5Hash(cacheKey)
	var cacheTtl types.Duration
	if config.CacheTimeouts.ContentItem != nil {
		cacheTtl = *config.CacheTimeouts.ContentItem
	} else {
		cacheTtl = internal.Config.CacheTimeouts.ContentItem
	}
	parsed, err := site.ParseTemplate("content-item", path, config, customContext, nocache, cacheKey, time.Duration(cacheTtl),
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			var results = new(types.ContentItemResult)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", time.Duration(cacheTtl), time.Duration(cacheTtl), func() ([]byte, error) {
				return api.ContentItemRaw(hostName, langId, slug, id, orfl, int64(relatedAmount), groupId,
					relatedTitleTranslated, relatedTitleTranslatedMinTermFreq, relatedTitleTranslatedMaxQueryTerms, relatedTitleTranslatedBoost,
					relatedTitle, relatedTitleMinTermFreq, relatedTitleMaxQueryTerms, relatedTitleBoost,
					relatedTags, relatedTagsMinTermFreq, relatedTagsMaxQueryTerms, relatedTagsBoost,
				)
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
				var args = make([]any, 0, 10)
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
			format := results.GetThumbFormat()
			results.ThumbFormat = format.Name
			results.ThumbsWidth = int32(format.Width)
			results.ThumbsHeight = int32(format.Height)
			results.ThumbsAmount = int32(format.Amount)
			results.ThumbRetina = format.Retina
			results.ThumbType = format.Type
			results.ThumbWidth = results.ThumbsWidth
			results.ThumbHeight = results.ThumbsHeight
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
			ctx["content_item"] = results
			ctx["related"] = results.Related
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
			log.Println("content item not found", slug, id, hostName, r.Header.Get("Referer"), r.Header.Get("User-Agent"), langId, ip, r.URL.String())
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

func getContentItemFunc(hostName string, config *types.Config, langId string, groupId int64, nocache bool) func(args ...any) *types.ContentItemResult {
	relatedTitleTranslated := config.Related.TitleTranslated
	if relatedTitleTranslated == nil {
		relatedTitleTranslated = internal.Config.Related.TitleTranslated
	}
	relatedTitleTranslatedMinTermFreq := config.Related.TitleTranslatedMinTermFreq
	if relatedTitleTranslatedMinTermFreq == nil {
		relatedTitleTranslatedMinTermFreq = internal.Config.Related.TitleTranslatedMinTermFreq
	}
	relatedTitleTranslatedMaxQueryTerms := config.Related.TitleTranslatedMaxQueryTerms
	if relatedTitleTranslatedMaxQueryTerms == nil {
		relatedTitleTranslatedMaxQueryTerms = internal.Config.Related.TitleTranslatedMaxQueryTerms
	}
	relatedTitleTranslatedBoost := config.Related.TitleTranslatedBoost
	if relatedTitleTranslatedBoost == nil {
		relatedTitleTranslatedBoost = internal.Config.Related.TitleTranslatedBoost
	}
	relatedTitle := config.Related.Title
	if relatedTitle == nil {
		relatedTitle = internal.Config.Related.Title
	}
	relatedTitleMinTermFreq := config.Related.TitleMinTermFreq
	if relatedTitleMinTermFreq == nil {
		relatedTitleMinTermFreq = internal.Config.Related.TitleMinTermFreq
	}
	relatedTitleMaxQueryTerms := config.Related.TitleMaxQueryTerms
	if relatedTitleMaxQueryTerms == nil {
		relatedTitleMaxQueryTerms = internal.Config.Related.TitleMaxQueryTerms
	}
	relatedTitleBoost := config.Related.TitleBoost
	if relatedTitleBoost == nil {
		relatedTitleBoost = internal.Config.Related.TitleBoost
	}
	relatedTags := config.Related.Tags
	if relatedTags == nil {
		relatedTags = internal.Config.Related.Tags
	}
	relatedTagsMinTermFreq := config.Related.TagsMinTermFreq
	if relatedTagsMinTermFreq == nil {
		relatedTagsMinTermFreq = internal.Config.Related.TagsMinTermFreq
	}
	relatedTagsMaxQueryTerms := config.Related.TagsMaxQueryTerms
	if relatedTagsMaxQueryTerms == nil {
		relatedTagsMaxQueryTerms = internal.Config.Related.TagsMaxQueryTerms
	}
	relatedTagsBoost := config.Related.TagsBoost
	if relatedTagsBoost == nil {
		relatedTagsBoost = internal.Config.Related.TagsBoost
	}
	var cacheTime types.Duration
	if config.CacheTimeouts.ContentItem != nil {
		cacheTime = *config.CacheTimeouts.ContentItem
	} else {
		cacheTime = internal.Config.CacheTimeouts.ContentItem
	}
	return func(args ...any) *types.ContentItemResult {
		parsingName := true
		var id int64
		var slug string
		orfl := !config.General.FakeVideoPage
		var relatedAmount = int64(config.General.ContentRelatedAmount)
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
			case "id":
				id, _ = strconv.ParseInt(val, 10, 64)
			case "slug":
				slug = val
			case "related_amount":
				relatedAmount, _ = strconv.ParseInt(val, 10, 64)
			case "orfl":
				orfl, _ = strconv.ParseBool(val)
			case "group_id":
				groupId, _ = strconv.ParseInt(val, 10, 32)
			case "cache":
				cacheTime = types.Duration(types.ParseHumanDuration(val))
			}
		}
		if id == 0 && slug == "" {
			log.Println("can't get content item: no id or slug", hostName)
			return nil
		}
		cacheKey := "content-item:" + helpers.Md5Hash(
			fmt.Sprintf("%s:%s:%d:%s:%v:%d:%d", hostName, langId, id, slug, orfl, relatedAmount, groupId),
		)
		results, err := db.GetCachedTimeout(cacheKey+":data", time.Duration(cacheTime), time.Duration(cacheTime), func() ([]byte, error) {
			return api.ContentItemRaw(hostName, langId, slug, id, orfl, relatedAmount, groupId,
				relatedTitleTranslated, relatedTitleTranslatedMinTermFreq, relatedTitleTranslatedMaxQueryTerms, relatedTitleTranslatedBoost,
				relatedTitle, relatedTitleMinTermFreq, relatedTitleMaxQueryTerms, relatedTitleBoost,
				relatedTags, relatedTagsMinTermFreq, relatedTagsMaxQueryTerms, relatedTagsBoost,
			)
		}, nocache)
		if err != nil {
			log.Println("can't get content item:", err, hostName)
			return nil
		}
		var result = new(types.ContentItemResult)
		err = json.Unmarshal(results, result)
		if err != nil {
			log.Println("can't get content item:", err, hostName)
			return nil
		}
		return result
	}
}
