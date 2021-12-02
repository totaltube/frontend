package api

import "sersh.com/totaltube/frontend/types"

func CountView(siteDomain string, params types.CountViewParams) (err error) {
	_, err = ApiRequest(siteDomain, methodPost, uriCountView, params)
	return
}
