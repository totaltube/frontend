package api

import "sersh.com/totaltube/frontend/types"

func CategoryClick(params types.CountClickParams) (err error) {
	_, err = apiRequest(methodPost, uriCategoryClick, params)
	return
}
