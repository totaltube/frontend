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

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var FakePlayer = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*types.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	slug := chi.URLParam(r, "slug")
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id == 0 && slug == "" {
		Output404(w, r, "content item not found")
		return
	}
	if id > 0 && config.Routes.IdXorKey > 0 {
		id = id ^ config.Routes.IdXorKey
	}
	orfl := !config.General.FakeVideoPage
	relatedAmount := config.General.ContentRelatedAmount
	customContext := generateCustomContext(w, r, "fake-player")
	params := customContext["params"].(map[string]string)
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	cacheKey := "fake-player:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s:%v:%d:%d", hostName, langId, id, slug, orfl, relatedAmount, groupId),
	)

	cacheTtl := time.Minute * 30
	parsed, err := site.ParseTemplate("fake-player", path, config, customContext, nocache, cacheKey, cacheTtl,
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
					url:  site.GetLink("content", config, hostName, langId, false, args...),
					code: 301,
				}
			}
			ctx["content_item"] = results
			ctx["related"] = results.Related
			// let's add first category name to params
			if len(results.Categories) > 0 {
				params["category"] = results.Categories[0].Slug
			} else {
				params["category"] = "default"
			}
			id := results.Id
			if config.Routes.IdXorKey > 0 {
				id = id ^ config.Routes.IdXorKey
			}
			params["id"] = strconv.FormatInt(id, 10)
			params["slug"] = results.Slug
			ctx["params"] = params
			return ctx, nil
		}, w, r)
	if err != nil {
		if rErr, ok := err.(redirectErr); ok {
			http.Redirect(w, r, rErr.url, rErr.code)
			if internal.Config.General.EnableAccessLog {
				log.Printf("Redirected to %s", rErr.url)
			}
			return
		}
		if strings.Contains(err.Error(), "not found") {
			Output404(w, r, err.Error())
			return
		}
		Output500(w, r, err)
		return
	}
	
	if middlewares.HeadersSent(w) {
		return
	}
	//w.Header().Set("X-Robots-Tag", "noindex")
	render.HTML(w, r, string(parsed))
})
