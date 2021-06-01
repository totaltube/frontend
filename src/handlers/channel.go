package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/segmentio/encoding/json"
	"log"
	"math/rand"
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
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	modelId, _ := strconv.ParseInt(c.Query(config.Params.ModelId), 10, 64)
	modelSlug := utils.ImmutableString(c.Query(config.Params.ModelSlug))
	categorySlug := utils.ImmutableString(c.Query(config.Params.CategorySlug))
	categoryId, _ := strconv.ParseInt(c.Query(config.Params.CategoryId), 10, 64)
	sortBy := utils.ImmutableString(c.Query(config.Params.SortBy, "dated"))
	sortByViewsTimeframe := utils.ImmutableString(c.Query(config.Params.SortByViewsTimeframe))
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
	channelId, _ := strconv.ParseInt(c.Params("id", c.Query(config.Params.ChannelId, "0")), 10, 64)
	channelSlug := utils.ImmutableString(c.Params("slug", c.Query(config.Params.ChannelSlug)))
	if channelId == 0 && channelSlug == "" {
		return Generate404(c, "channel not found")
	}
	durationGte, _ := strconv.ParseInt(c.Query(config.Params.DurationGte, "0"), 10, 64)
	durationLt, _ := strconv.ParseInt(c.Query(config.Params.DurationLt, "0"), 10, 64)
	customContext := generateCustomContext(c, "channel")
	cacheKey := "channel:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d:%d:%s",
			hostName, langId, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationGte, durationLt, categoryId, categorySlug),
	)
	cacheTtl := time.Minute * 15
	ip := utils.ImmutableString(c.IP())
	userAgent := utils.ImmutableString(c.Get("User-Agent"))
	parsed, err := site.ParseTemplate("channel", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			channelInfoCacheKey := fmt.Sprintf("in:chinfo:%d:%s:%s", channelId, channelSlug, langId)
			channelInfoCacheTtl := time.Hour * 24 + time.Duration(rand.Intn(3600*6))*time.Second
			channelInfoCached, err := db.GetCachedTimeout(channelInfoCacheKey, channelInfoCacheTtl, time.Hour*4, func() ([]byte, error) {
				_, rawResponse, err := api.ChannelInfo(hostName, langId, channelId, channelSlug)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			channelInfo := new(types.ChannelResult)
			err = json.Unmarshal(channelInfoCached, channelInfo)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			var results *types.ContentResults
			results, _, err = api.Content(hostName, api.ContentParams{
				Ip:           net.ParseIP(ip),
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
				DurationGte:  durationGte,
				DurationLt:   durationLt,
				UserAgent:    userAgent,
			})
			if err != nil {
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
		if strings.Contains(err.Error(), "not found") {
			return Generate404(c, err.Error())
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}
