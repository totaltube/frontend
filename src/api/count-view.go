package api

import "sersh.com/totaltube/frontend/types"

func CountView(siteConfig *types.Config, params types.CountViewParams) (err error) {
	_, err = Request(siteConfig, methodPost, uriCountView, params)
	return
}
