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

func Channel(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	modelId, _ := strconv.ParseInt(c.Query(config.Params.ModelId), 10, 64)
	modelSlug := c.Query(config.Params.ModelSlug)
	categorySlug := c.Query(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(c.Query(config.Params.CategoryId), 10, 64)
	sortBy := c.Query(config.Params.SortBy, "dated")
	sortByViewsTimeframe := c.Query(config.Params.SortByViewsTimeframe)
	channelId, _ := strconv.ParseInt(c.Params("id", c.Query(config.Params.ChannelId, "0")), 10, 64)
	channelSlug := c.Params("slug", c.Query(config.Params.ChannelSlug))
	if channelId == 0 && channelSlug == "" {
		return Generate404(c, "channel not found")
	}
	durationFrom, _ := strconv.ParseInt(c.Query(config.Params.DurationFrom, "0"), 10, 64)
	durationTo, _ := strconv.ParseInt(c.Query(config.Params.DurationTo, "0"), 10, 64)
	customContext := generateCustomContext(c, "channel")
	cacheKey := "channel:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%d:%s:%s:%s:%d:%d:%s:%d:%d:%d:%s",
			langId, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug),
	)
	cacheTtl := time.Minute * 15
	parsed, err := site.ParseTemplate("channel", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			channelInfoCacheKey := fmt.Sprintf("chinfo:%d:%s:%s", channelId, channelSlug, langId)
			channelInfoCacheTtl := time.Minute * 60 * 24
			channelInfoCached := db.GetCached(channelInfoCacheKey)
			var channelInfo *types.ChannelResult
			if channelInfoCached != nil && !nocache {
				channelInfo = new(types.ChannelResult)
				err := json.Unmarshal(channelInfoCached, channelInfo)
				if err != nil {
					log.Println(err)
					return ctx, err
				}
			} else {
				var err error
				channelInfo, err = api.ChannelInfo(langId, channelId, channelSlug)
				if err != nil {
					if strings.Contains(err.Error(), "not found") {
						_ = Generate404(c, "channel not found")
						return ctx, types.ErrResponseSent
					}
					log.Println(err)
					return ctx, err
				}
				err = db.PutCached(channelInfoCacheKey, helpers.ToJSON(channelInfo), channelInfoCacheTtl)
				if err != nil {
					log.Println(err)
					return ctx, err
				}
			}
			var results *types.ContentResults
			var err error
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
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					_ = Generate404(c, err.Error())
					return ctx, types.ErrResponseSent
				}
				return ctx, err
			}
			ctx["channel"] = channelInfo
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
