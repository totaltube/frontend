package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var RedirectToContentItem = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*types.Config)
	langId := r.Context().Value("lang").(string)
	hostName := r.Context().Value("hostName").(string)
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	slug := r.URL.Query().Get("slug")
	if id <= 0 && slug == "" {
		Output404(w, r, "page not found")
		return
	}
	results, err := api.ContentItem(hostName, langId, slug, id, true, 0, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			Output404(w, r, "content item not found")
			return
		}
		Output500(w, r, err)
		return
	}
	link := site.GetLink("content-item", config, hostName, langId, false, "slug", results.Slug, "id", results.Id, "categories", results.Categories)
	http.Redirect(w, r, link, http.StatusFound)
	if internal.Config.General.EnableAccessLog {
		log.Println(hostName, http.StatusFound, link)
	}
})
