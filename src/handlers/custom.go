package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
)

var Custom = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value(types.ContextKeyPath).(string)
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	templateName := r.Context().Value(types.ContextKeyCustomTemplateName).(string)
	page, _ := strconv.ParseInt(helpers.FirstNotEmpty(chi.URLParam(r, "page"), r.URL.Query().Get(config.Params.Page), "1"), 10, 16)
	if page <= 0 {
		page = 1
	}
	customContext := generateCustomContext(w, r, "custom/"+templateName)
	customContext["page"] = page
	parsed, err := site.ParseCustomTemplate(templateName, path, config, customContext, nocache, w, r)
	if err != nil {
		if err1, ok := err.(site.ErrSendResponse); ok {
			if err1.Headers != nil {
				for key := range err1.Headers {
					w.Header().Set(key, err1.Headers.Get(key))
				}
			}
			if err1.Redirect != "" {
				if err1.Code != 301 {
					err1.Code = 302
				}
				http.Redirect(w, r, err1.Redirect, err1.Code)
				if internal.Config.General.EnableAccessLog {
					log.Println(hostName, err1.Code, err1.Redirect)
				}
				return
			}
			if err1.JSON != nil {
				render.JSON(w, r, err1.JSON)
				return
			}
			if err1.Data != nil {
				if err1.Code != 0 {
					w.WriteHeader(err1.Code)
				}
				_, _ = w.Write(err1.Data)
				return
			}
			if err1.Text != "" {
				if err1.Code != 0 {
					w.WriteHeader(err1.Code)
				}
				render.HTML(w, r, err1.Text)
				return
			}
		}
		if strings.Contains(err.Error(), "not found") {
			log.Println(hostName, err)
			Output404(w, r, err.Error())
			return
		}
		log.Println(hostName, err)
		Output500(w, r, err)
		return
	}
	if middlewares.HeadersSent(w) {
		return
	}
	render.HTML(w, r, string(parsed))
})
