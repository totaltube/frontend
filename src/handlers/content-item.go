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
	"strings"
	"time"
)

func ContentItem(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	slug := c.Params("slug")
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if id == 0 && slug == "" {
		return Generate404(c, "content item not found")
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	customContext := generateCustomContext(c, "content-item")
	cacheKey := "content-item:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%d:%s:%v:%d", langId, id, slug, orfl, relatedAmount),
	)
	cacheTtl := time.Minute * 30
	parsed, err := site.ParseTemplate("content-item", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			var results *types.ContentItemResult
			var err error
			results, err = api.ContentItem(langId, slug, id, orfl, int64(relatedAmount))
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					_ = Generate404(c, "content item not found")
					return ctx, types.ErrResponseSent
				}
				return ctx, err
			}
			ctx["content_item"] = results
			ctx["related"] = results.Related
			return ctx, nil
		})
	if err != nil {
		if err == types.ErrResponseSent {
			return nil
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}
