package api

import "sersh.com/totaltube/frontend/types"

func CountView(params types.CountViewParams) (err error) {
	_, err = apiRequest(methodPost, uriCountView, params)
	return
}
