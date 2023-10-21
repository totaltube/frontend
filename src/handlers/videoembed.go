package handlers

import (
	"encoding/json"
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
	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var VideoEmbed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "video-embed")
	params := customContext["params"].(map[string]string)
	cacheKey := "video-embed:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s", hostName, langId, id, slug),
	)
	cacheTtl := time.Minute * 30
	parsed, err := site.ParseTemplate("video-embed", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			// getting category information from cache or from api
			var results = new(types.ContentItemResult)
			var err error
			var response json.RawMessage
			response, err = db.GetCachedTimeout(cacheKey+":data", cacheTtl, cacheTtl, func() ([]byte, error) {
				return api.ContentItemRaw(hostName, langId, slug, id, true, 0, groupId)
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
			if results.Type != "video" {
				log.Println("wrong type of video on embed page - ", results.Type)
				return ctx, errors.New("content item not found")
			}
			ctx["content_item"] = results
			ctx["related"] = results.Related
			// let's add first category name to params
			if len(results.Categories) > 0 {
				params["category"] = results.Categories[0].Slug
			} else {
				params["category"] = "default"
			}
			ctx["params"] = params
			return ctx, nil
		}, w, r)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			render.Status(r, 404)
			render.HTML(w, r, "content not found")
			return
		}
		Output500(w, r, err)
		return
	}
	w.Header().Set("X-Robots-Tag", "noindex")
	render.HTML(w, r, string(parsed))
})
