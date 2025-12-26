package handlers

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/types"
)

var Rating = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	ip := r.Context().Value(types.ContextKeyIp).(string)
	slug, _ := url.PathUnescape(chi.URLParam(r, "slug"))
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id == 0 && slug == "" {
		Output404(w, r, "content item not found")
		return
	}
	if id > 0 && config.Routes.IdXorKey > 0 {
		id = id ^ config.Routes.IdXorKey
	}
	var like bool
	if r.URL.Query().Get(config.Params.Like) != "" {
		like, _ = strconv.ParseBool(r.URL.Query().Get(config.Params.Like))
	}
	returnFunc := func() {
		// Function which return json at the end.
		render.JSON(w, r, M{"success": true})
	}
	if botDetector.IsBot(r.Header.Get("User-Agent")) {
		// Do not count anything for bots
		returnFunc()
		return
	}
	// All calculations are done in background
	go func() {
		err := api.Rating(config, ip, id, slug, like)
		if err != nil {
			log.Println(err)
		}
	}()
	returnFunc()
})
