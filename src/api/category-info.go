package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func CategoryInfo(siteDomain, lang string, categoryId int64, categorySlug string) (result *types.CategoryResult, err error) {
	var response json.RawMessage
	response, err = apiRequest(siteDomain, methodGet, uriCategoryInfo, url.Values{
		"id":   []string{strconv.FormatInt(categoryId, 10)},
		"slug": []string{categorySlug},
		"lang": []string{lang},
	})
	if err != nil {
		return
	}
	result = new(types.CategoryResult)
	err = json.Unmarshal(response, result)
	return
}
