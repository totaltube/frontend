package api

import "sersh.com/totaltube/frontend/types"

func TopContentClick(params types.CountClickParams) (err error) {
	_, err = apiRequest(methodPost, uriTopContentClick, params)
	return
}
