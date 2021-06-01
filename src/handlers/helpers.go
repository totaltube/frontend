package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"net/url"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"strconv"
	"strings"
	"time"
)

func generateCustomContext(c *fiber.Ctx, templateName string) pongo2.Context {
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	params := helpers.FiberAllParams(c)
	query := helpers.FiberAllQuery(c)
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
	queryString := string(c.Context().QueryArgs().QueryString())
	headers := map[string]string{}
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[utils.ImmutableString(string(key))] = utils.ImmutableString(string(value))
	})
	cookies := map[string]string{}
	c.Request().Header.VisitAllCookie(func(key, value []byte) {
		cookies[utils.ImmutableString(string(key))] = utils.ImmutableString(string(value))
	})
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
	default:
		if r, ok := config.Custom[strings.TrimPrefix(templateName, "custom/")]; ok {
			route = r
		}
	}
	switch templateName {
	case "category", "model", "channel", "top-content", "popular", "new", "long", "search":
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
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	var globals = make(map[string]interface{})
	customContext := pongo2.Context{
		"page_template":   templateName,
		"lang":            internal.GetLanguage(langId),
		"ip":              utils.ImmutableString(c.IP()),
		"nocache":         nocache,
		"languages":       internal.GetLanguages(),
		"page":            page,
		"host":            hostName,
		"params":          params,
		"query":           query,
		"querystring":     queryString,
		"headers":         headers,
		"cookies":         cookies,
		"canonical_query": canonicalQuery,
		"config":          config,
		"global_config":   internal.Config,
		"route":           route,
	}
	customContext["set_cookie"] = func(name string, value interface{}, expire interface{}) {
		var expires = time.Now().Add(time.Minute * 60)
		if e, ok := expire.(time.Time); ok {
			expires = e
		}
		if e, ok := expire.(time.Duration); ok {
			expires = time.Now().Add(e)
		}
		if e, ok := expire.(int64); ok {
			expires = time.Now().Add(time.Hour * 24 * time.Duration(e))
		}
		if e, ok := expire.(int); ok {
			expires = time.Now().Add(time.Hour * 24 * time.Duration(e))
		}
		var cookie = &fiber.Cookie{
			Name:    name,
			Value:   fmt.Sprintf("%v", value),
			Expires: expires,
		}
		c.Cookie(cookie)
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

func Generate404(c *fiber.Ctx, errMessage string) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	customContext := generateCustomContext(c, "404")
	customContext["error"] = errMessage
	cacheKey := fmt.Sprintf("404:%s", langId)
	cacheTtl := time.Minute * 5
	parsed, err := site.ParseTemplate("404", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			return ctx, nil
		})
	if err != nil {
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Status(fiber.StatusNotFound).Send(parsed)
}

func Generate500(c *fiber.Ctx, errMessage string) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	customContext := generateCustomContext(c, "500")
	customContext["error"] = errMessage
	cacheKey := fmt.Sprintf("500:%s", langId)
	cacheTtl := time.Minute * 5
	parsed, err := site.ParseTemplate("500", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			return ctx, nil
		})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html")
	return c.Status(fiber.StatusInternalServerError).Send(parsed)
}
