package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"net"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"time"
)

func New(c *fiber.Ctx) error {
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
	modelSlug := c.Query(config.Params.ModelSlug)
	categorySlug := c.Query(config.Params.CategorySlug)
	categoryId, _ := strconv.ParseInt(c.Query(config.Params.CategoryId), 10, 64)
	sortBy := "dated"
	channelId, _ := strconv.ParseInt(c.Query(config.Params.ChannelId, "0"), 10, 64)
	channelSlug := c.Query(config.Params.ChannelSlug)
	durationFrom, _ := strconv.ParseInt(c.Query(config.Params.DurationGte, "0"), 10, 64)
	durationTo, _ := strconv.ParseInt(c.Query(config.Params.DurationLt, "0"), 10, 64)
	customContext := generateCustomContext(c, "new")
	cacheKey := "new:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%d:%d:%s:%d:%d:%d:%s",
			hostName, langId, page, channelSlug, channelId,
			modelId, modelSlug, durationFrom, durationTo, categoryId, categorySlug),
	)
	cacheTtl := time.Minute * 15
	parsed, err := site.ParseTemplate("new", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
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
				DurationGte:  durationFrom,
				DurationLt:   durationTo,
				UserAgent:    c.Get("User-Agent"),
			})
			if err != nil {
				return ctx, err
			}
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
