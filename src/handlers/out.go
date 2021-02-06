package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/logocomune/botdetector"
	"log"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
)

var botDetector = botdetector.New()

type countInfo struct {
	config     *site.Config
	categoryId string
	contentId  string
	ip string
	countType string
	countThumbId int64
}

var countChannel = make(chan countInfo, 100)

func Out(c *fiber.Ctx) error {
	config := c.Locals("config").(*site.Config)
	ip := c.IP()
	redirectUrl := helpers.DecryptBase64(c.Query(config.Params.CountRedirect))
	countType := c.Query(config.Params.CountType)
	countThumbId, _ := strconv.ParseInt(c.Query(config.Params.CountThumbId, "-1"), 10, 16)
	returnFunc := func() error {
		// Function which redirects or return json at the end.
		if redirectUrl != "" {
			if internal.Config.General.Nginx && strings.HasPrefix(redirectUrl, "/") {
				c.Set("X-Accel-Redirect", redirectUrl)
				return c.Send([]byte(""))
			}
			return c.Redirect(redirectUrl)
		}
		return c.JSON(M{"success": true})
	}
	if botDetector.IsBot(c.Get("User-Agent")) {
		// Do not count anything for bots
		return returnFunc()
	}
	// All calculations are done in background
	categoryIdParam := c.Query(config.Params.CategoryId)
	contentIdParam := c.Query(config.Params.ContentId)
	info := countInfo{
		config:       config,
		categoryId:   categoryIdParam,
		contentId:    contentIdParam,
		ip:           ip,
		countType:    countType,
		countThumbId: countThumbId,
	}
	// lets count in background in separate goroutine, by sending in buffered channel
	countChannel <- info
	return returnFunc()
}

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
			db.SessMutex.Lock(info.ip)
			defer db.SessMutex.Unlock(info.ip)
			sess := db.GetSession(info.ip)
			defer db.SaveSession(info.ip, sess)
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
				err := api.CountView(types.CountViewParams{
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
				err := api.TopCategoriesClick(types.CountClickParams{
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
				err := api.TopContentClick(types.CountClickParams{
					Ip: info.ip,
					Id: countId,
				})
				log.Println("top content click api error:", err)
				return
			case info.config.Params.CountTypeCategory:
				if sess.LastClickType == info.countType && sess.LastClickId == countId {
					return
				}
				sess.LastClickType = info.countType
				sess.LastClickId = countId
				categoryId, _ := strconv.ParseInt(info.categoryId, 10, 32)
				err := api.CategoryClick(categoryId, types.CountClickParams{
					Ip: info.ip,
					Id: countId,
				})
				log.Println("category click api error: ", err)
				return
			}
		}()
	}
}
