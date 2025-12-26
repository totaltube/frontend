package api

import "sersh.com/totaltube/frontend/types"

func TopContentClick(siteConfig *types.Config, params types.CountClickParams) (err error) {
	_, err = Request(siteConfig, methodPost, uriTopContentClick, params)
	return
}
