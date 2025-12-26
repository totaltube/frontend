package api

import (
	"log"

	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

func Rating(
	siteConfig *types.Config, ip string, id int64, slug string, isLike bool,
) (err error) {
	likes := 0
	dislikes := 0
	if isLike {
		likes = 1
	} else {
		dislikes = 1
	}
	if internal.Config.General.EnableAccessLog {
		siteName := ""
		if siteConfig != nil {
			siteName = siteConfig.Hostname
		}
		log.Println("Rating", siteName, ip, id, slug, likes, dislikes)
	}
	_, err = Request(siteConfig, methodPost, uriRating, M{
		"ip":       ip,
		"id":       id,
		"slug":     slug,
		"likes":    likes,
		"dislikes": dislikes,
	})
	return
}
