package api

import "sersh.com/totaltube/frontend/types"

func TopCategoriesClick(siteDomain string, params types.CountClickParams) (err error) {
	_, err = apiRequest(siteDomain, methodPost, uriTopCategoriesClick, params)
	return
}
