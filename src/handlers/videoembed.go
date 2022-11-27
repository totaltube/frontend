package handlers

import (
	"fmt"
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
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var VideoEmbed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	ip := r.Context().Value("ip").(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	customContext := generateCustomContext(w, r, "video-embed")
	cacheKey := "video-embed:" + helpers.Md5Hash(
		fmt.Sprintf("%s:%s:%d:%s", hostName, langId, id, slug),
	)
	cacheTtl := time.Minute * 30
	parsed, err := site.ParseTemplate("video-embed", path, config, customContext, nocache, cacheKey, cacheTtl,
		func(ctx pongo2.Context) (pongo2.Context, error) {
			// getting category information from cache or from api
			var results *types.ContentItemResult
			var err error
			results, err = api.ContentItem(hostName, langId, slug, id, true, 0, groupId)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return ctx, errors.New("content item not found")
				}
				return ctx, err
			}
			if results.Type != "video" {
				return ctx, errors.New("content item not found")
			}
			ctx["content_item"] = results
			ctx["related"] = results.Related
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
