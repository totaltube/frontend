package handlers

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"net/url"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/site"
	"strconv"
	"strings"
	"time"
)

func TopCategories(c *fiber.Ctx) error {
	path := c.Locals("path").(string)
	config := c.Locals("config").(*site.Config)
	hostName := c.Locals("hostName").(string)
	langId := c.Locals("lang").(string)
	if ref := c.Get("Referer"); ref != "" && !config.General.DisableCategoriesRedirect {
		if u, err := url.Parse(ref); err == nil &&
			strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") != hostName &&
			!botDetector.IsBot(c.Get("User-Agent")) {
			var s = strings.ToLower(u.Path + " " + u.RawQuery)
			if categories, err := db.GetCachedTopCategories(hostName); err == nil {
				for _, cat := range categories.Items {
					for _, t := range cat.Tags {
						if strings.Contains(s, t) {
							var redirectUrl = config.Routes.Category
							if config.General.MultiLanguage {
								redirectUrl = strings.ReplaceAll(config.Routes.LanguageTemplate, ":lang", langId)
								redirectUrl = strings.ReplaceAll(redirectUrl, ":route", config.Routes.Category)
							}
							redirectUrl = strings.ReplaceAll(redirectUrl, ":slug", cat.Slug)
							redirectUrl = strings.ReplaceAll(redirectUrl, ":id", strconv.FormatInt(int64(cat.Id), 10))
							return c.Redirect(redirectUrl)
						}
					}
				}
			}
		}
	}
	nocache, _ := strconv.ParseBool(c.Query(config.Params.Nocache, "false"))
	page, _ := strconv.ParseInt(c.Params("page", c.Query(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(c, "top-categories")
	cacheKey := fmt.Sprintf("top-categories:%s:%d", langId, page)
	cacheTtl := time.Second * 5
	if page > 1 {
		cacheTtl = time.Minute * 5
	}
	parsed, err := site.ParseTemplate("top-categories", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			results, err := api.TopCategories(hostName, langId, page)
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
