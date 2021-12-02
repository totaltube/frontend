package handlers

import (
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
	sortBy := utils.ImmutableString(c.Params(":sort", c.Query(config.Params.SortBy, "title")))
	query := utils.ImmutableString(c.Query(config.Params.SearchQuery))
	amount := config.General.ModelsPerPage
	customContext := generateCustomContext(c, "models")
	cacheKey := "models:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%s:%d",
			hostName, langId, page, sortBy, query, amount),
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
		if strings.Contains(err.Error(), "not found") {
			return Generate404(c, err.Error())
		}
		return err
	}
	c.Set("Content-Type", "text/html")
	return c.Send(parsed)
}

func getModelsListFunc(hostName string, langId string, defaultAmount int64) func(args ...interface{}) *types.ModelResults {
	return func(args ...interface{}) *types.ModelResults {
		parsingName := true
		var amount = defaultAmount
		var page int64
		var sortBy = api.SortTitle
		var searchQuery = ""
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
				page, _ = strconv.ParseInt(val, 10, 64)
			case "sort":
				sortBy = api.SortBy(val)
			case "amount":
				amount, _ = strconv.ParseInt(val, 10, 64)
			case "search_query":
				searchQuery = val
			}
		}
		results, _, err := api.ModelsList(hostName, langId, page, sortBy, amount, searchQuery)
		if err != nil {
			log.Println("can't get models list:", err)
			return nil
		}
		return results
	}
}
