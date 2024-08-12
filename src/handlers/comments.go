package handlers

import (
	"github.com/go-chi/render"
	"net/http"
)

var Comments = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	/*config := r.Context().Value("config").(*types.Config)
	hostName := r.Context().Value("hostName").(string)
	ip := r.Context().Value("ip").(string)
	slug, _ := url.PathUnescape(chi.URLParam(r, "slug"))
	*/
	render.JSON(w, r, M{"success": true})
	return
})
