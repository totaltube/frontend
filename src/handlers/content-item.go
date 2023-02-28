package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

type redirectErr struct {
	code int
	url  string
}

func (r redirectErr) Error() string {
	return fmt.Sprintf("%d redirect to %s", r.code, r.url)
}

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
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "content-item")
	cacheKey := "content-item:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%v:%d:%d", hostName, langId, id, slug, orfl, relatedAmount, groupId),
	)
	cacheTtl := time.Minute * 60
	parsed, err := site.ParseTemplate("content-item", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			var results = new(types.ContentItemResult)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", cacheTtl, cacheTtl, func() ([]byte, error) {
				return api.ContentItemRaw(hostName, langId, slug, id, orfl, int64(relatedAmount), groupId)
			}, nocache)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return ctx, errors.New("content item not found")
				}
				return ctx, err
			}
			err = json.Unmarshal(response, results)
			if err != nil {
				return ctx, err
			}
			if slug != "" && results.Slug != slug {
				// Rephrased title and slug, need to make a 301 redirect
				categorySlug := "default"
				if len(results.Categories) > 0 {
					categorySlug = results.Categories[0].Slug
				}
				var args = make([]interface{}, 0, 10)
				args = append(args,
					"slug", results.Slug,
					"id", results.Id,
					"category", categorySlug,
				)
				for k := range r.URL.Query() {
					args = append(args, k, r.URL.Query().Get(k))
				}
				return nil, redirectErr{
					url: site.GetLink("content", config, "content-item", langId, args...),
					code: 301,
				}
			}
			format := results.GetThumbFormat()
			results.ThumbFormat = format.Name
			results.ThumbsWidth = int32(format.Width)
			results.ThumbsHeight = int32(format.Height)
			results.ThumbsAmount = int32(format.Amount)
			results.ThumbRetina = format.Retina
			results.ThumbType = format.Type
			results.ThumbWidth = results.ThumbsHeight
			results.ThumbHeight = results.ThumbsHeight
			ctx["content_item"] = results
			ctx["related"] = results.Related
			return ctx, nil
		}, w, r)
	if err != nil {
		if rErr, ok := err.(redirectErr); ok {
			http.Redirect(w, r, rErr.url, rErr.code)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			Output404(w, r, err.Error())
			return
		}
		Output500(w, r, err)
		return
	}
	render.HTML(w, r, string(parsed))
})

func getContentItemFunc(hostName string, langId string, groupId int64) func(args ...interface{}) *types.ContentItemResult {
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
			case "group_id":
				groupId, _ = strconv.ParseInt(val, 10, 32)
			}
		}
		if id == 0 && slug == "" {
			log.Println("can't get content item: no id or slug")
			return nil
		}
		results, err := api.ContentItem(hostName, langId, slug, id, orfl, relatedAmount, groupId)
		if err != nil {
			log.Println("can't get content item:", err)
			return nil
		}
		return results
	}
}
