package api

import (
	"log"

	"sersh.com/totaltube/frontend/internal"
)

func Rating(
	siteDomain, ip string, id int64, slug string, isLike bool,
) (err error) {
	likes := 0
	dislikes := 0
	if isLike {
		likes = 1
	} else {
		dislikes = 1
	}
	if internal.Config.General.EnableAccessLog {
		log.Println("Rating", siteDomain, ip, id, slug, likes, dislikes)
	}
	_, err = Request(siteDomain, methodPost, uriRating, M{
		"ip":       ip,
		"id":       id,
		"slug":     slug,
		"likes":    likes,
		"dislikes": dislikes,
	})
	return
}
