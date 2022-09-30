package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/logocomune/botdetector"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var botDetector = botdetector.New()

type countInfo struct {
	hostName     string
	config       *site.Config
	categoryId   string
	contentId    string
	ip           string
	countType    string
	countThumbId int64
}

var countChannel = make(chan countInfo, 100)

var Out = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*site.Config)
	hostName := r.Context().Value("hostName").(string)
	ip := r.Context().Value("ip").(string)
	redirectUrl := r.URL.Query().Get(config.Params.CountRedirect)
	encryptedRedirectUrl := r.URL.Query().Get("e" + config.Params.CountRedirect)
	if redirectUrl == "" && encryptedRedirectUrl != "" {
		redirectUrl = helpers.DecryptBase64(encryptedRedirectUrl)
	}
	countType := r.URL.Query().Get(config.Params.CountType)
	countThumbId, _ := strconv.ParseInt(helpers.FirstNotEmpty(r.URL.Query().Get(config.Params.CountThumbId), "-1"), 10, 16)
	returnFunc := func() {
		// Function which redirects or return json at the end.
		if redirectUrl != "" {
			if internal.Config.General.Nginx && strings.HasPrefix(redirectUrl, "/") {
				w.Header().Set("X-Accel-Redirect", redirectUrl)
				return
			}
			http.Redirect(w, r, redirectUrl, 302)
			return
		}
		render.JSON(w, r, M{"success": true})
		return
	}
	if botDetector.IsBot(r.Header.Get("User-Agent")) {
		// Do not count anything for bots
		returnFunc()
		return
	}
	// All calculations are done in background
	categoryIdParam := r.URL.Query().Get(config.Params.CategoryId)
	contentIdParam := r.URL.Query().Get(config.Params.ContentId)
	info := countInfo{
		hostName:     hostName,
		config:       config,
		categoryId:   categoryIdParam,
		contentId:    contentIdParam,
		ip:           ip,
		countType:    countType,
		countThumbId: countThumbId,
	}
	// let's count in background in separate goroutine, by sending in buffered channel
	countChannel <- info
	returnFunc()
})

func doCount() {
	// function to count in separate goroutine
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("recover in doCount:", r)
				}
			}()
			info := <-countChannel
			var ip = info.ip
			sess := db.GetSession(ip)
			defer db.SaveSession(ip, sess)
			var countId int64
			switch info.countType {
			case info.config.Params.CountTypeTopCategories:
				countId, _ = strconv.ParseInt(info.categoryId, 10, 64)
			case info.config.Params.CountTypeCategory, info.config.Params.CountTypeTopContent:
				countId, _ = strconv.ParseInt(info.contentId, 10, 64)
				if sess.LastViewType == info.countType && sess.LastViewId == countId {
					// no need to count view or click of this content
					return
				}
				sess.LastViewType = info.countType
				sess.LastViewId = countId
				// Let's count view of this content
				err := api.CountView(info.hostName, types.CountViewParams{
					Type:    "content",
					Id:      countId,
					Ip:      info.ip,
					ThumbId: int16(info.countThumbId),
				})
				if err != nil {
					log.Println("error counting view:", err)
					return
				}
			default:
				log.Println("wrong count type - " + info.countType)
				return
			}
			// now let's count click
			switch info.countType {
			case info.config.Params.CountTypeTopCategories:
				if sess.LastClickType == info.countType && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType
				sess.LastClickId = countId
				err := api.TopCategoriesClick(info.hostName, types.CountClickParams{
					Ip: info.ip,
					Id: countId,
				})
				if err != nil {
					log.Println("top categories click api error:", err)
				}
				return
			case info.config.Params.CountTypeTopContent:
				if sess.LastClickType == info.countType && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType
				sess.LastClickId = countId
				err := api.TopContentClick(info.hostName, types.CountClickParams{
					Ip: info.ip,
					Id: countId,
				})
				if err != nil {
					log.Println("top content click api error:", err)
				}
				return
			case info.config.Params.CountTypeCategory:
				if sess.LastClickType == info.countType && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType
				sess.LastClickId = countId
				categoryId, _ := strconv.ParseInt(info.categoryId, 10, 32)
				err := api.CategoryClick(info.hostName, categoryId, types.CountClickParams{
					Ip: info.ip,
					Id: countId,
				})
				if err != nil {
					log.Println("category click api error: ", err, info.hostName, categoryId, info.ip, countId)
				}
				return
			}
		}()
	}
}
