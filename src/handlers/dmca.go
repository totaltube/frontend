package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/render"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)


var Dmca = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.Context().Value("path").(string)
	config := r.Context().Value("config").(*site.Config)
	hostName := r.Context().Value("hostName").(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	langId := r.Context().Value("lang").(string)
	customContext := generateCustomContext(w,r, "dmca")
	cacheTtl := time.Minute * 15
	isOk := false
	var ip = r.Context().Value("ip").(string)
	session := db.GetSession(ip)
	defer db.SaveSession(ip, session)
	if session.LastDmca.IsZero() || session.LastDmca.Before(time.Now().Add(-time.Minute)) {
		session.DmcaAmount = 0
		session.LastDmca = time.Now()
	}
	if r.Method == "POST" {
		session.DmcaAmount++
		params := types.DmcaParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			render.JSON(w, r, M{"success": false, "value": "wrong parameters: "+err.Error()})
			return
		}
		isWhitelisted := false
		for _, e := range internal.Config.Frontend.CaptchaWhiteList {
			if e == params.Email {
				isWhitelisted = true
				break
			}
		}
		if session.DmcaAmount > internal.Config.Frontend.MaxDmcaMinute && !isWhitelisted {
			if params.CaptchaResponse == "" {
				render.JSON(w, r, M{"success": false, "value": "need captcha"})
				return
			}
			response := helpers.Fetch("https://hcaptcha.com/siteverify").
				WithFormData(M{
					"secret":   internal.Config.Frontend.CaptchaSecret,
					"response": params.CaptchaResponse,
				}).Json()
			verifyOk := false
			if success, ok := response["success"].(bool); ok && success {
				if h, ok := response["hostname"].(string);
					ok && strings.TrimPrefix(h, "www.") ==
						strings.TrimPrefix(strings.Split(r.URL.Hostname(), ":")[0], "www.") {
					verifyOk = true
				} else {
					log.Println("wrong hostname for hCaptcha!", strings.Split(r.URL.Hostname(), ":")[0], response["hostname"])
				}
			}
			if !verifyOk {
				render.JSON(w, r, M{"success": false, "value": "captcha error"})
				return
			}
		}
		err = api.Dmca(hostName, params)
		if err != nil {
			render.JSON(w, r, M{"success": false, "value": err.Error()})
			return
		}
		accepted := render.GetAcceptedContentType(r)
		if accepted == render.ContentTypeJSON {
			render.JSON(w, r, M{"success": true})
			return
		} else {
			isOk = true
		}
		nocache = true
	}
	renderCaptcha := session.DmcaAmount > internal.Config.Frontend.MaxDmcaMinute
	cacheKey := "dmca:" + hostName + ":" + langId + ":" + strconv.FormatBool(isOk) + ":" + strconv.FormatBool(renderCaptcha)
	parsed, err := site.ParseTemplate("dmca", path, config, customContext, nocache, cacheKey, cacheTtl,
		func() (pongo2.Context, error) {
			ctx := pongo2.Context{}
			ctx["ok"] = isOk
			ctx["render_captcha"] = renderCaptcha
			return ctx, nil
		}, w,r)
	if err != nil {
		render.JSON(w, r, M{"success": false, "value": err.Error()})
		return
	}
	render.HTML(w, r, string(parsed))
})