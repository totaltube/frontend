package handlers

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mileusna/useragent"
	"github.com/samber/lo"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

func generateCustomContext(_ http.ResponseWriter, r *http.Request, templateName string) pongo2.Context {
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	langId := r.Context().Value(types.ContextKeyLang).(string)
	refreshTranslations := r.URL.Query().Get(config.Params.Nocache) == "3"
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	var params = make(map[string]string)
	urlParams := chi.RouteContext(r.Context()).URLParams
	for k := range urlParams.Keys {
		params[urlParams.Keys[k]] = urlParams.Values[k]
	}
	var query = make(map[string]string)
	for k := range r.URL.Query() {
		query[k] = r.URL.Query().Get(k)
	}
	uri := r.URL.Path
	userAgent := r.Header.Get("User-Agent")
	var changedQuery = make(map[string]string)
	for k, v := range query {
		if k == config.Params.SortBy {
			k = "sort_by"
			if v == config.Params.SortByRand {
				v = "rand"
			} else if v == config.Params.SortByViews {
				v = "views"
			} else if v == config.Params.SortByDuration {
				v = "duration"
			} else if v == config.Params.SortByDate {
				v = "dated"
			}
			changedQuery[k] = v
		} else if k == config.Params.SortByViewsTimeframe {
			k = "timeframe"
			changedQuery[k] = v
		} else if k == config.Params.ChannelSlug {
			k = "channel_slug"
			changedQuery[k] = v
		} else if k == config.Params.ChannelId {
			k = "channel_id"
			changedQuery[k] = v
		} else if k == config.Params.CategorySlug {
			k = "category_slug"
			changedQuery[k] = v
		} else if k == config.Params.CategoryId {
			k = "category_id"
			changedQuery[k] = v
		} else if k == config.Params.ModelSlug {
			k = "model_slug"
			changedQuery[k] = v
		} else if k == config.Params.ModelId {
			k = "model_id"
			changedQuery[k] = v
		} else if k == config.Params.Page {
			k = "page"
			changedQuery[k] = v
		} else if k == config.Params.ContentSlug {
			k = "content_slug"
			changedQuery[k] = v
		} else if k == config.Params.ContentId {
			k = "content_id"
			changedQuery[k] = v
		} else if k == config.Params.SearchQuery {
			k = "search_query"
			changedQuery[k] = v
		} else if k == config.Params.DurationGte {
			k = "duration_gte"
			changedQuery[k] = v
		} else if k == config.Params.DurationLt {
			k = "duration_lt"
			changedQuery[k] = v
		}
	}
	for k, v := range changedQuery {
		query[k] = v
	}
	queryString := r.URL.RawQuery
	headers := make(map[string]string)
	headers["Host"] = r.Host
	for k := range r.Header {
		headers[k] = r.Header.Get(k)
	}
	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	canonicalQuery := url.Values{}
	route := config.Routes.TopCategories
	switch templateName {
	case "top-categories":
		route = config.Routes.TopCategories
	case "category":
		route = config.Routes.Category
	case "model":
		route = config.Routes.Model
	case "channel":
		route = config.Routes.Channel
	case "top-content":
		route = config.Routes.TopContent
	case "popular":
		route = config.Routes.Popular
	case "new":
		route = config.Routes.New
	case "long":
		route = config.Routes.Long
	case "search":
		route = config.Routes.Search
	case "models":
		route = config.Routes.Models
	case "content-item":
		route = config.Routes.ContentItem
	case "fake-player":
		route = config.Routes.FakePlayer
	case "video-embed":
		route = config.Routes.VideoEmbed
	case "sitemap-video":
		route = config.Sitemap.Route
	default:
		if rr, ok := config.Routes.Custom[strings.TrimPrefix(templateName, "custom/")]; ok {
			route = rr
		}
	}
	if lo.Contains([]string{"category", "model", "channel", "top-content", "popular", "new", "long", "search"}, templateName) || strings.HasPrefix(templateName, "custom/") {
		if categorySlug, ok := query[config.Params.CategorySlug]; ok {
			canonicalQuery.Set(config.Params.CategorySlug, categorySlug)
			if templateName == "category" {
				if _, ok := params["slug"]; ok {
					canonicalQuery.Del(config.Params.CategorySlug)
				}
				if _, ok := params["id"]; ok {
					canonicalQuery.Del(config.Params.CategorySlug)
				}
			}
		}
		if categoryId, ok := query[config.Params.CategoryId]; ok {
			canonicalQuery.Set(config.Params.CategoryId, categoryId)
			if templateName == "category" {
				if _, ok := params["slug"]; ok {
					canonicalQuery.Del(config.Params.CategoryId)
				}
				if _, ok := params["id"]; ok {
					canonicalQuery.Del(config.Params.CategoryId)
				}
			}
		}
		if channelSlug, ok := query[config.Params.ChannelSlug]; ok {
			canonicalQuery.Set(config.Params.ChannelSlug, channelSlug)
			if templateName == "channel" {
				if _, ok := params["slug"]; ok {
					canonicalQuery.Del(config.Params.ChannelSlug)
				}
				if _, ok := params["id"]; ok {
					canonicalQuery.Del(config.Params.ChannelSlug)
				}
			}
		}
		if channelId, ok := query[config.Params.ChannelId]; ok {
			canonicalQuery.Set(config.Params.ChannelId, channelId)
			if templateName == "channel" {
				if _, ok := params["slug"]; ok {
					canonicalQuery.Del(config.Params.ChannelId)
				}
				if _, ok := params["id"]; ok {
					canonicalQuery.Del(config.Params.ChannelId)
				}
			}
		}
		if modelSlug, ok := query[config.Params.ModelSlug]; ok {
			canonicalQuery.Set(config.Params.ModelSlug, modelSlug)
			if templateName == "model" {
				if _, ok := params["slug"]; ok {
					canonicalQuery.Del(config.Params.ModelSlug)
				}
				if _, ok := params["id"]; ok {
					canonicalQuery.Del(config.Params.ModelSlug)
				}
			}
		}
		if modelId, ok := query[config.Params.ModelId]; ok {
			canonicalQuery.Set(config.Params.ModelId, modelId)
			if templateName == "model" {
				if _, ok := params["slug"]; ok {
					canonicalQuery.Del(config.Params.ModelId)
				}
				if _, ok := params["id"]; ok {
					canonicalQuery.Del(config.Params.ModelId)
				}
			}
		}
		if durationFrom, ok := query[config.Params.DurationGte]; ok {
			canonicalQuery.Set(config.Params.DurationGte, durationFrom)
		}
		if durationTo, ok := query[config.Params.DurationLt]; ok {
			canonicalQuery.Set(config.Params.DurationLt, durationTo)
		}
		if searchQuery, ok := query[config.Params.SearchQuery]; ok {
			canonicalQuery.Set(config.Params.SearchQuery, searchQuery)
			if templateName == "search" {
				if _, ok := params["query"]; ok {
					canonicalQuery.Del(config.Params.SearchQuery)
				}
			}
		}
		if sortBy, ok := query[config.Params.SortBy]; ok &&
			templateName != "popular" && templateName != "new" && templateName != "long" {
			canonicalQuery.Set(config.Params.SortBy, sortBy)
			if sortBy == config.Params.SortByViews {
				if sortTimeframe, ok := query[config.Params.SortByViewsTimeframe]; ok {
					canonicalQuery.Set(config.Params.SortByViewsTimeframe, sortTimeframe)
				}
			}
		}
	}
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	var globals = make(map[string]interface{})
	ip := r.Context().Value(types.ContextKeyIp).(string)
	var countryGroup = internal.DetectCountryGroup(net.ParseIP(ip))
	groupId := countryGroup.Id
	customContext := pongo2.Context{
		"page_template":       templateName,
		"lang":                internal.GetLanguage(langId),
		"ip":                  ip,
		"uri":                 uri,
		"user_agent":          userAgent,
		"nocache":             nocache,
		"languages":           internal.GetLanguages(config),
		"page":                page,
		"host":                hostName,
		"params":              params,
		"query":               query,
		"querystring":         queryString,
		"headers":             headers,
		"cookies":             cookies,
		"canonical_query":     canonicalQuery,
		"config":              config,
		"global_config":       internal.Config,
		"route":               route,
		"country_group":       countryGroup,
		"group_id":            countryGroup.Id,
		"refreshTranslations": refreshTranslations,
		"parse_ua": func(ua ...string) useragent.UserAgent {
			if len(ua) > 0 {
				return useragent.Parse(ua[0])
			}
			return useragent.Parse(r.UserAgent())
		},
		"get_content":         getContentFunc(hostName, langId, userAgent, ip, groupId),
		"get_top_content":     getTopContentFunc(hostName, langId, groupId, config),
		"get_top_categories":  getTopCategoriesFunc(hostName, langId, groupId, config),
		"get_content_item":    getContentItemFunc(hostName, config, langId, groupId, nocache),
		"get_models_list":     getModelsListFunc(hostName, langId, int64(config.General.ModelsPerPage), groupId),
		"get_categories_list": getCategoriesListFunc(hostName, langId, 100, groupId),
		"get_channels_list":   getChannelsListFunc(hostName, langId, 100, groupId),
		"get_category_top":    getCategoryTopFunc(hostName, langId, groupId, config),
		"get_category":        getCategoryFunc(hostName, langId),
		"get_model":           getModelFunc(hostName, langId, groupId),
		"xor_id": func(id *pongo2.Value) int64 {
			idInt := int64(id.Integer())
			if idInt > 0 && config.Routes.IdXorKey > 0 {
				return idInt ^ config.Routes.IdXorKey
			}
			return idInt
		},
		"add_random_content": func(items []*types.ContentResult, amount ...interface{}) []*types.ContentResult {
			var amt int64 = 0
			if len(amount) > 0 {
				amt, _ = strconv.ParseInt(fmt.Sprintf("%v", amount[0]), 10, 64)
			}
			if amt == 0 {
				amt = int64(internal.Config.Options.Popularity.Layouts.Category.Amount)
			}
			if int(amt) <= len(items) {
				return items
			}
			results, _, err := api.Content(hostName, api.ContentParams{
				Ip:        net.ParseIP(ip),
				UserAgent: userAgent,
				Lang:      langId,
				Amount:    amt - int64(len(items)),
				Sort:      api.SortRandNoPaging,
			})
			if err != nil {
				log.Println(err)
				return items
			}
			return append(items, results.Items...)
		},
		"merge": func(dst, src interface{}) interface{} {
			dv := reflect.ValueOf(dst)
			sv := reflect.ValueOf(src)
			dv2 := reflect.AppendSlice(dv, sv)
			return dv2.Interface()
		},
	}
	// Functions to set and get vars, which will be saved between calls.
	customContext["set_var"] = func(name string, value interface{}) {
		globals[name] = value
	}
	customContext["get_var"] = func(name string) interface{} {
		return globals[name]
	}
	return customContext
}

func Output404(w http.ResponseWriter, r *http.Request, errMessage string) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	langId := r.Context().Value(types.ContextKeyLang).(string)
	customContext := generateCustomContext(w, r, "404")
	customContext["error"] = errMessage
	cacheKey := fmt.Sprintf("404:%s:%s:%s", hostName, langId, helpers.Md5Hash(errMessage))
	cacheTtl := time.Minute * 5
	parsed, err := site.ParseTemplate("404", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			return ctx, nil
		}, w, r)
	if err != nil {
		panic(err)
	}
	render.Status(r, 404)
	render.HTML(w, r, string(parsed))
}

func Output500(w http.ResponseWriter, r *http.Request, err error) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	langId := r.Context().Value(types.ContextKeyLang).(string)
	customContext := generateCustomContext(w, r, "404")
	customContext["error"] = err.Error()
	log.Println(err, hostName, langId)
	cacheKey := fmt.Sprintf("500:%s:%s:%s", hostName, langId, helpers.Md5Hash(err.Error()))
	cacheTtl := time.Minute * 5
	var parsed []byte
	parsed, err = site.ParseTemplate("500", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			return ctx, nil
		}, w, r)
	if err != nil {
		panic(err)
	}
	render.Status(r, 500)
	render.HTML(w, r, string(parsed))
}
