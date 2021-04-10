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

func Model(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	modelId, _ := strconv.ParseInt(c.Params("id", c.Query(config.Params.ModelId)), 10, 64)
	modelSlug := c.Params("slug", c.Query(config.Params.ModelSlug))
	if modelId == 0 && modelSlug == "" {
		return Generate404(c, "model not found")
	}
	categorySlug := c.Query(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(c.Query(config.Params.CategoryId), 10, 64)
	sortBy := c.Query(config.Params.SortBy, "dated")
	sortByViewsTimeframe := c.Query(config.Params.SortByViewsTimeframe)
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
	channelId, _ := strconv.ParseInt(c.Query(config.Params.ChannelId, "0"), 10, 64)
	channelSlug := c.Query(config.Params.ChannelSlug)
	durationFrom, _ := strconv.ParseInt(c.Query(config.Params.DurationGte, "0"), 10, 64)
	durationTo, _ := strconv.ParseInt(c.Query(config.Params.DurationLt, "0"), 10, 64)
	customContext := generateCustomContext(c, "model")
	cacheKey := "model:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%d:%s:%s:%s:%d:%d:%s:%d:%d:%d:%s",
			langId, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug),
	)
	cacheTtl := time.Minute * 15
	parsed, err := site.ParseTemplate("model", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			modelInfoCacheKey := fmt.Sprintf("minfo:%d:%s:%s", modelId, modelSlug, langId)
			modelInfoCacheTtl := time.Minute * 60 * 24
			modelInfoCached := db.GetCached(modelInfoCacheKey)
			var modelInfo *types.ModelResult
			if modelInfoCached != nil && !nocache {
				modelInfo = new(types.ModelResult)
				err := json.Unmarshal(modelInfoCached, modelInfo)
				if err != nil {
					log.Println(err)
					return ctx, err
				}
			} else {
				var err error
				modelInfo, err = api.ModelInfo(hostName, langId, modelId, modelSlug)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						_ = Generate404(c, "model not found")
						return ctx, types.ErrResponseSent
					}
					log.Println(err)
					return ctx, err
				}
				err = db.PutCached(modelInfoCacheKey, helpers.ToJSON(modelInfo), modelInfoCacheTtl)
				if err != nil {
					log.Println(err)
					return ctx, err
				}
			}
			var results *types.ContentResults
			var err error
			results, _, err = api.Content(hostName, api.ContentParams{
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
				UserAgent:    c.Get("User-Agent"),
			})
			if err != nil {
				return ctx, err
			}
			ctx["model"] = modelInfo
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
