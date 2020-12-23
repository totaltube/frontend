package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"net/url"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"strconv"
	"time"
)

func TopCategories(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	route := config.Routes.TopCategories
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	customContext := pongo2.Context{
		"page_template":   "top-categories",
		"lang":            internal.GetLanguage(langId),
		"languages":       internal.GetLanguages(),
		"page":            page,
		"params":          helpers.FiberAllParams(c),
		"query":           helpers.FiberAllQuery(c),
		"canonical_query": url.Values{},
		"config":          config,
		"route":           route,
	}
	cacheKey := fmt.Sprintf("top-categories:%s:%d", langId, page)
	cacheTtl := time.Second * 5
	if page > 1 {
		cacheTtl = time.Minute * 5
	}
	parsed, err := site.ParseTemplate("top-categories", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			results, err := api.TopCategories(langId, page)
			if err != nil {
				return ctx, err
			}
			ctx["top_categories"] = results
			ctx["total"] = int64(results.Total)
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
