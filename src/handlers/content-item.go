package handlers

import (
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"log"
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
	hostName := c.Locals("hostName").(string)
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	langId := c.Locals("lang").(string)
	slug := utils.ImmutableString(c.Params("slug"))
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if id == 0 && slug == "" {
		return Generate404(c, "content item not found")
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	customContext := generateCustomContext(c, "content-item")
	cacheKey := "content-item:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%v:%d", hostName, langId, id, slug, orfl, relatedAmount),
	)
	cacheTtl := time.Minute * 60
	parsed, err := site.ParseTemplate("content-item", path, config, customContext, nocache, cacheKey, cacheTtl,
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

func getContentItemFunc(hostName string, langId string) func(args ...interface{}) *types.ContentItemResult {
	return func(args ...interface{}) *types.ContentItemResult {
		parsingName := true
		var id int64
		var slug string
		var orfl bool
		var relatedAmount int64
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
			case "id":
				id, _ = strconv.ParseInt(val, 10, 64)
			case "slug":
				slug = val
			case "related_amount":
				relatedAmount, _ = strconv.ParseInt(val, 10, 64)
			case "orfl":
				orfl, _ = strconv.ParseBool(val)
			}
		}
		if id == 0 && slug == "" {
			log.Println("can't get content item: no id or slug")
			return nil
		}
		results, err := api.ContentItem(hostName, langId, slug, id, orfl, relatedAmount)
		if err != nil {
			log.Println("can't get content item:", err)
			return nil
		}
		return results
	}
}
