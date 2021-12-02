package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func CategoryInfo(siteDomain, lang string, categoryId int64, categorySlug string) (result *types.CategoryResult, rawResponse json.RawMessage, err error) {
	rawResponse, err = ApiRequest(siteDomain, methodGet, uriCategoryInfo, url.Values{
		"id":   []string{strconv.FormatInt(categoryId, 10)},
		"slug": []string{categorySlug},
		"lang": []string{lang},
	})
	if err != nil {
		return
	}
	result = new(types.CategoryResult)
	err = json.Unmarshal(rawResponse, result)
	return
}
