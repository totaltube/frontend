package api

import "sersh.com/totaltube/frontend/types"

func TopCategoriesClick(siteConfig *types.Config, params types.CountClickParams) (err error) {
	_, err = Request(siteConfig, methodPost, uriTopCategoriesClick, params)
	return
}
