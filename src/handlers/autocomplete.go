package handlers

import (
	"log"
	"net/http"
	"sersh.com/totaltube/frontend/types"

	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
)

var Autocomplete = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*types.Config)
	langId := r.Context().Value("lang").(string)
	hostName := r.Context().Value("hostName").(string)
	searchQuery := r.URL.Query().Get(config.Params.SearchQuery)
	results, err := api.Autocomplete(hostName, searchQuery, langId, config)
	if err != nil {
		log.Println("Error querying autocomplete api: ", err)
		render.JSON(w, r, A{})
		return
	}
	render.JSON(w, r, results.Items)
})