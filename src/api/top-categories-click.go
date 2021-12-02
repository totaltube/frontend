package api

import "sersh.com/totaltube/frontend/types"

func TopCategoriesClick(siteDomain string, params types.CountClickParams) (err error) {
	_, err = ApiRequest(siteDomain, methodPost, uriTopCategoriesClick, params)
	return
}
