package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

func FakePlayer(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	slug := utils.CopyString(c.Params("slug"))
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if id == 0 && slug == "" {
		return Generate404(c, "content item not found")
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	customContext := generateCustomContext(c, "fake-player")
	cacheKey := "fake-player:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%v:%d", hostName, langId, id, slug, orfl, relatedAmount),
	)
	cacheTtl := time.Minute * 30
	parsed, err := site.ParseTemplate("fake-player", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			var results *types.ContentItemResult
			var err error
			results, err = api.ContentItem(hostName, langId, slug, id, orfl, int64(relatedAmount))
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return ctx, errors.New("content item not found")
				}
				return ctx, err
			}
			ctx["content_item"] = results
			ctx["related"] = results.Related
			return ctx, nil
		}, c)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return Generate404(c, err.Error())
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	c.Set("X-Robots-Tag", "noindex")
	return c.Send(parsed)
}
