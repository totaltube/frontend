package api

import (
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func CategoryClick(siteConfig *types.Config, categoryId int64, params types.CountClickParams) (err error) {
	uriParams := url.Values{}
	uriParams.Set("category_id", strconv.FormatInt(categoryId, 10))
	_, err = Request(siteConfig, methodPost, uriCategoryClick+ApiUri("?"+uriParams.Encode()), params)
	return
}
