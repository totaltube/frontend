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

func Category(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	categorySlug := utils.ImmutableString(c.Params("slug", c.Query(config.Params.CategorySlug)))
	categoryId, _ := strconv.ParseInt(c.Params("id", c.Query(config.Params.CategoryId)), 10, 64)
	if categoryId == 0 && categorySlug == "" {
		return Generate404(c, "category not found")
	}
	sortBy := utils.ImmutableString(c.Query(config.Params.SortBy))
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
	sortByViewsTimeframe := utils.ImmutableString(c.Query(config.Params.SortByViewsTimeframe))
	channelId, _ := strconv.ParseInt(c.Query(config.Params.ChannelId, "0"), 10, 64)
	channelSlug := utils.ImmutableString(c.Query(config.Params.ChannelSlug))
	modelId, _ := strconv.ParseInt(c.Query(config.Params.ModelId, "0"), 10, 64)
	modelSlug := utils.ImmutableString(c.Query(config.Params.ModelSlug))
	durationFrom, _ := strconv.ParseInt(c.Query(config.Params.DurationGte, "0"), 10, 64)
	durationTo, _ := strconv.ParseInt(c.Query(config.Params.DurationLt, "0"), 10, 64)
	customContext := generateCustomContext(c, "category")
	cacheKey := "category:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%d:%s:%s:%s:%d:%d:%s:%d:%d",
			hostName, langId, categoryId, categorySlug, page, sortBy, sortByViewsTimeframe, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo),
	)
	filtered := channelId > 0 || channelSlug != "" || modelId > 0 || modelSlug != "" || sortBy != "" ||
		durationTo > 0 || durationFrom > 0
	cacheTtl := time.Second * 5
	if page > 1 || filtered {
		cacheTtl = time.Minute * 5
	}
	ip := utils.ImmutableString(c.IP())
	userAgent := c.Get("User-Agent")
	parsed, err := site.ParseTemplate("category", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			categoryInfoCacheKey := fmt.Sprintf("in:cinfo:%d:%s:%s", categoryId, categorySlug, langId)
			categoryInfoCacheTtl := time.Hour*24 + time.Duration(rand.Intn(3600*6))*time.Second
			categoryInfoCached, err := db.GetCachedTimeout(categoryInfoCacheKey, categoryInfoCacheTtl, time.Hour*4, func() ([]byte, error) {
				_, rawResponse, err := api.CategoryInfo(hostName, langId, categoryId, categorySlug)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			categoryInfo := new(types.CategoryResult)
			err = json.Unmarshal(categoryInfoCached, categoryInfo)
			if err != nil {
				log.Println(err)
				return ctx, err
			}
			var results *types.ContentResults
			if filtered {
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
					DurationGte:  durationFrom,
					DurationLt:   durationTo,
					UserAgent:    userAgent,
				})
			} else {
				ctx["count"] = true
				results, err = api.Category(hostName, langId, categoryId, categorySlug, page)
			}
			if err != nil {
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
		if strings.Contains(err.Error(), "not found") {
			return Generate404(c, err.Error())
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}

func getCategoryTopFunc(hostName string, langId string) func(args ...interface{}) *types.ContentResults {
	return func(args ...interface{}) *types.ContentResults {
		parsingName := true
		var categoryId int64
		var categorySlug string
		var page int64
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
				page, _ = strconv.ParseInt(val, 10, 16)
			case "category_id", "id":
				categoryId, _ = strconv.ParseInt(val, 10, 32)
			case "category_slug", "slug":
				categorySlug = val
			}
		}
		if categoryId == 0 && categorySlug == "" {
			log.Println("error getting top category content - need to set category_id or category_slug param")
			return nil
		}
		if results, err := api.Category(hostName, langId, categoryId, categorySlug, page); err != nil {
			log.Println("error getting category top content: ", err)
			return nil
		} else {
			return results
		}
	}
}
