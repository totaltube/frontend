package handlers

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/logocomune/botdetector"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

var botDetector = botdetector.New()

type countInfo struct {
	hostName     string
	categoryId   int64
	contentId    int64
	countView    bool
	ip           string
	countType    types.CountType
	countThumbId int64
}

var countChannel = make(chan countInfo, 1000)

var Out = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value("config").(*types.Config)
	hostName := r.Context().Value("hostName").(string)
	ip := r.Context().Value("ip").(string)
	redirectUrl := r.URL.Query().Get(config.Params.CountRedirect)
	encryptedRedirectUrl := r.URL.Query().Get("e" + config.Params.CountRedirect)
	if redirectUrl == "" && encryptedRedirectUrl != "" {
		redirectUrl = helpers.DecryptBase64(encryptedRedirectUrl)
	}
	countTypeParam := r.URL.Query().Get(config.Params.CountType)
	countThumbId, _ := strconv.ParseInt(helpers.FirstNotEmpty(r.URL.Query().Get(config.Params.CountThumbId), "-1"), 10, 16)
	returnFunc := func() {
		// Function which redirects or return json at the end.
		if redirectUrl != "" {
			if internal.Config.General.Nginx && strings.HasPrefix(redirectUrl, "/") {
				w.Header().Set("X-Accel-Redirect", redirectUrl)
				return
			}
			http.Redirect(w, r, redirectUrl, http.StatusFound)
			if internal.Config.General.EnableAccessLog {
				log.Println(hostName, 302, redirectUrl)
			}
			return
		}
		render.JSON(w, r, M{"success": true})
	}
	if botDetector.IsBot(r.Header.Get("User-Agent")) {
		// Do not count anything for bots
		returnFunc()
		return
	}
	// All calculations are done in background
	categoryId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.CategoryId), 10, 32)
	if categoryId > 0 && config.Routes.IdXorKey > 0 {
		categoryId = categoryId ^ config.Routes.IdXorKey
	}
	contentId, _ := strconv.ParseInt(r.URL.Query().Get(config.Params.ContentId), 10, 64)
	if contentId > 0 && config.Routes.IdXorKey > 0 {
		contentId = contentId ^ config.Routes.IdXorKey
	}
	// by default, we count also view on corresponding category or content, but if you set not to count the view, it will not be counted
	countView := true
	if len(r.URL.Query().Get(config.Params.CountView)) > 0 {
		countView, _ = strconv.ParseBool(r.URL.Query().Get(config.Params.CountView))
	}
	countType := types.CountTypeNone
	if countTypeParam == config.Params.CountTypeCategory {
		countType = types.CountTypeCategory
	} else if countTypeParam == config.Params.CountTypeTopCategories {
		countType = types.CountTypeTopCategories
	} else if countTypeParam == config.Params.CountTypeTopContent {
		countType = types.CountTypeTopContent
	} else if countTypeParam == config.Params.CountTypeCategoryView {
		countType = types.CountTypeCategoryView
	}
	info := countInfo{
		hostName:     hostName,
		categoryId:   categoryId,
		contentId:    contentId,
		ip:           ip,
		countType:    countType,
		countThumbId: countThumbId,
		countView:    countView,
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
			groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
			var countId int64
			if info.countView {
				switch info.countType {
				case types.CountTypeTopCategories, types.CountTypeCategoryView:
					countId = info.categoryId
				default:
					countId = info.contentId
					if sess.LastViewType == info.countType.String() && sess.LastViewId == countId {
						// no need to count view or click of this content
						return
					}
					sess.LastViewType = info.countType.String()
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
				}
			}
			// now let's count click
			switch info.countType {
			case types.CountTypeTopCategories:
				if sess.LastClickType == info.countType.String() && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType.String()
				sess.LastClickId = countId
				err := api.TopCategoriesClick(info.hostName, types.CountClickParams{
					Ip:      info.ip,
					Id:      countId,
					GroupId: groupId,
				})
				if err != nil {
					log.Println("top categories click api error:", err)
				}
				return
			case types.CountTypeTopContent:
				if sess.LastClickType == info.countType.String() && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType.String()
				sess.LastClickId = countId
				err := api.TopContentClick(info.hostName, types.CountClickParams{
					Ip:      info.ip,
					Id:      countId,
					GroupId: groupId,
				})
				if err != nil {
					log.Println("top content click api error:", err)
				}
				return
			case types.CountTypeCategory:
				if sess.LastClickType == info.countType.String() && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType.String()
				sess.LastClickId = countId
				err := api.CategoryClick(info.hostName, info.categoryId, types.CountClickParams{
					Ip:      info.ip,
					Id:      countId,
					GroupId: groupId,
				})
				if err != nil {
					log.Println("category click api error: ", err, info.hostName, info.categoryId, info.ip, countId)
				}
				return
			}
		}()
	}
}
