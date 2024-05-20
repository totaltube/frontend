package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log"
	"net/http"
	"net/url"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

var Rating = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*types.Config)
	hostName := r.Context().Value("hostName").(string)
	ip := r.Context().Value("ip").(string)
	slug, _ := url.PathUnescape(chi.URLParam(r, "slug"))
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if id == 0 && slug == "" {
		Output404(w, r, "content item not found")
		return
	}
	var like bool
	if r.URL.Query().Get(config.Params.Like) != "" {
		like, _ = strconv.ParseBool(r.URL.Query().Get(config.Params.Like))
	}
	returnFunc := func() {
		// Function which return json at the end.
		render.JSON(w, r, M{"success": true})
		return
	}
	if botDetector.IsBot(r.Header.Get("User-Agent")) {
		// Do not count anything for bots
		returnFunc()
		return
	}
	// All calculations are done in background
	go func() {
		err := api.Rating(hostName, ip, id, slug, like)
		if err != nil {
			log.Println(err)
		}
	}()
	returnFunc()
})
