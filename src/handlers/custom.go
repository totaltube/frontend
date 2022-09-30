package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
)

var Custom = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*site.Config)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	templateName := r.Context().Value("custom_template_name").(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r,"page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(w, r, "custom/"+templateName)
	customContext["page"] = page
	parsed, err := site.ParseCustomTemplate(templateName, path, config, customContext, nocache, w, r)
	if err != nil {
		if err1, ok := err.(site.ErrSendResponse); ok {
			if err1.Redirect != "" {
				if err1.RedirectCode != 301 {
					err1.RedirectCode = 302
				}
				http.Redirect(w, r, err1.Redirect, err1.RedirectCode)
				return
			}
			if err1.JSON != nil {
				render.JSON(w, r, err1.JSON)
				return
			}
			if err1.Text != nil {
				render.HTML(w, r, string(err1.Text))
				return
			}
		}
		Output500(w, r, err)
		return
	}
	render.HTML(w, r, string(parsed))
})
