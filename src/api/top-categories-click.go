package api

import "sersh.com/totaltube/frontend/types"

func TopCategoriesClick(params types.CountClickParams) (err error) {
	_, err = apiRequest(methodPost, uriTopCategoriesClick, params)
	return
}
