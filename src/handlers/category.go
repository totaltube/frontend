package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/encoding/json"
	"log"
	"net"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
	"time"
)

func Category(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	categorySlug := c.Params("slug", c.Query(config.Params.CategorySlug))
	categoryId, _ := strconv.ParseInt(c.Params("id", c.Query(config.Params.CategoryId)), 10, 64)
	if categoryId == 0 && categorySlug == "" {
		return Generate404(c, "category not found")
	}
	sortBy := c.Query(config.Params.SortBy)
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
	sortByViewsTimeframe := c.Query(config.Params.SortByViewsTimeframe)
	channelId, _ := strconv.ParseInt(c.Query(config.Params.ChannelId, "0"), 10, 64)
	channelSlug := c.Query(config.Params.ChannelSlug)
	modelId, _ := strconv.ParseInt(c.Query(config.Params.ModelId, "0"), 10, 64)
	modelSlug := c.Query(config.Params.ModelSlug)
	durationFrom, _ := strconv.ParseInt(c.Query(config.Params.DurationGte, "0"), 10, 64)
	durationTo, _ := strconv.ParseInt(c.Query(config.Params.DurationLt, "0"), 10, 64)
	customContext := generateCustomContext(c, "category")
	cacheKey := "category:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%d:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d",
			langId, categoryId, categorySlug, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo),
	)
	filtered := channelId > 0 || channelSlug != "" || modelId > 0 || modelSlug != "" || sortBy != "" ||
		durationTo > 0 || durationFrom > 0
	cacheTtl := time.Second * 5
	if page > 1 || filtered {
		cacheTtl = time.Minute * 5
	}
	parsed, err := site.ParseTemplate("category", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			categoryInfoCacheKey := fmt.Sprintf("cinfo:%d:%s:%s", categoryId, categorySlug, langId)
			categoryInfoCacheTtl := time.Minute * 60 * 24
			categoryInfoCached := db.GetCached(categoryInfoCacheKey)
			var categoryInfo *types.CategoryResult
			if categoryInfoCached != nil && !nocache {
				categoryInfo = new(types.CategoryResult)
				err := json.Unmarshal(categoryInfoCached, categoryInfo)
				if err != nil {
					log.Println(err)
					return ctx, err
				}
			} else {
				var err error
				categoryInfo, err = api.CategoryInfo(langId, categoryId, categorySlug)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						_ = Generate404(c, "category not found")
						return ctx, types.ErrResponseSent
					}
					log.Println(err)
					return ctx, err
				}
				err = db.PutCached(categoryInfoCacheKey, helpers.ToJSON(categoryInfo), categoryInfoCacheTtl)
				if err != nil {
					log.Println(err)
					return ctx, err
				}
			}
			var results *types.ContentResults
			var err error
			if filtered {
				results, err = api.Content(api.ContentParams{
					Ip:           net.ParseIP(c.IP()),
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
				})
			} else {
				ctx["count"] = true
				results, err = api.Category(langId, categoryId, categorySlug, page)
			}
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					_ = Generate404(c, err.Error())
					return nil, types.ErrResponseSent
				}
				return ctx, err
			}
			ctx["category"] = categoryInfo
			ctx["content"] = results
			ctx["total"] = results.Total
			ctx["from"] = int64(results.From)
			ctx["to"] = int64(results.To)
			ctx["page"] = int64(results.Page)
			ctx["pages"] = int64(results.Pages)
			return ctx, nil
		})
	if err != nil {
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}
