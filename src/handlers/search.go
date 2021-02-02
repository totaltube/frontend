package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"net"
	"net/url"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"time"
)

func Search(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	searchQuery, _ := url.PathUnescape(c.Params("query"))
	if searchQuery == "" {
		searchQuery = c.Query(config.Params.SearchQuery)
	}
	if searchQuery == "" {
		return errors.New("search query not set")
	}
	isNatural, _ := strconv.ParseBool(config.Params.SearchNatural)
	modelId, _ := strconv.ParseInt(c.Query(config.Params.ModelId), 10, 64)
	modelSlug := c.Query(config.Params.ModelSlug)
	categorySlug := c.Query(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(c.Query(config.Params.CategoryId), 10, 64)
	sortBy := c.Query(config.Params.SortBy)
	sortByTimeframe := c.Query(config.Params.SortByViewsTimeframe)
	channelId, _ := strconv.ParseInt(c.Query(config.Params.ChannelId, "0"), 10, 64)
	channelSlug := c.Query(config.Params.ChannelSlug)
	durationFrom, _ := strconv.ParseInt(c.Query(config.Params.DurationFrom, "0"), 10, 64)
	durationTo, _ := strconv.ParseInt(c.Query(config.Params.DurationTo, "0"), 10, 64)
	customContext := generateCustomContext(c, "search")
	cacheKey := "search:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%d:%s:%d:%d:%s:%d:%d:%d:%s:%s:%s",
			langId, page, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug, sortBy, searchQuery),
	)
	cacheTtl := time.Minute * 15
	parsed, err := site.ParseTemplate("search", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			var results *types.ContentResults
			var err error
			results, err = api.Content(api.ContentParams{
				Ip:           net.ParseIP(c.IP()),
				SearchQuery:  searchQuery,
				IsNatural:    isNatural,
				Lang:         langId,
				Page:         page,
				CategoryId:   categoryId,
				CategorySlug: categorySlug,
				ChannelId:    channelId,
				ChannelSlug:  channelSlug,
				ModelId:      modelId,
				ModelSlug:    modelSlug,
				Sort:         api.SortBy(sortBy),
				Timeframe:    sortByTimeframe,
				DurationGte:  durationFrom,
				DurationLt:   durationTo,
			})
			if err != nil {
				return ctx, err
			}
			ctx["search_query"] = searchQuery
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
