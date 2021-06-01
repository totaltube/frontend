package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"log"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
	"time"
)

func TopContent(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(c, "top-content")
	cacheKey := fmt.Sprintf("top-content:%s:%s:%d", hostName, langId, page)
	cacheTtl := time.Second * 5
	if page > 1 {
		cacheTtl = time.Minute * 5
	}
	parsed, err := site.ParseTemplate("top-content", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			var results *types.ContentResults
			var err error
			results, err = api.TopContent(hostName, langId, page)
			if err != nil {
				log.Println(err)
				return ctx, nil
			}
			if page == 1 {
				ctx["count"] = true
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
		if strings.Contains(err.Error(), "not found") {
			return Generate404(c, err.Error())
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}
