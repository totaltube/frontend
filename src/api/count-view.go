package api

import "sersh.com/totaltube/frontend/types"

func CountView(siteDomain string, params types.CountViewParams) (err error) {
	_, err = apiRequest(siteDomain, methodPost, uriCountView, params)
	return
}
