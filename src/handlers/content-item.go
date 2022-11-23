package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var ContentItem = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*site.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	slug := chi.URLParam(r, "slug")
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id == 0 && slug == "" {
		Output404(w, r, "content item not found")
		return
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	customContext := generateCustomContext(w, r, "content-item")
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
		}, w, r)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			Output404(w, r, err.Error())
			return
		}
		Output500(w, r, err)
		return
	}
	render.HTML(w, r, string(parsed))
})

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
