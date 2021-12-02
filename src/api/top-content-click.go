package api

import "sersh.com/totaltube/frontend/types"

func TopContentClick(siteDomain string, params types.CountClickParams) (err error) {
	_, err = ApiRequest(siteDomain, methodPost, uriTopContentClick, params)
	return
}
