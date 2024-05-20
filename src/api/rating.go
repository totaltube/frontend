package api

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
	_, err = Request(siteDomain, methodPost, uriRating, M{
		"ip":       ip,
		"id":       id,
		"slug":     slug,
		"likes":    likes,
		"dislikes": dislikes,
	})
	return
}
