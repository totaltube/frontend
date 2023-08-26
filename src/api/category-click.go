package api

import (
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func CategoryClick(siteDomain string, categoryId int64, params types.CountClickParams) (err error) {
	uriParams := url.Values{}
	uriParams.Set("category_id", strconv.FormatInt(categoryId, 10))
	_, err = Request(siteDomain, methodPost, uriCategoryClick+ApiUri("?"+uriParams.Encode()), params)
	return
}
