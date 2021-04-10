package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"time"
)

func Models(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	// can be title, total, popular
	sortBy := c.Params(":sort", c.Query(config.Params.SortBy, "title"))
	query := c.Query(config.Params.SearchQuery)
	amount := config.General.ModelsPerPage
	customContext := generateCustomContext(c, "models")
	cacheKey := "models:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%d:%s:%s:%d",
			langId, page, sortBy, query, amount),
	)
	cacheTtl := time.Minute * 15
	parsed, err := site.ParseTemplate("models", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			var results *types.ModelResults
			var err error
			results, _, err = api.ModelsList(hostName, langId, page, api.SortBy(sortBy), int64(amount), query)
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
