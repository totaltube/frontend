package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func CategoryInfo(siteConfig *types.Config, lang string, categoryId int64, categorySlug string) (result *types.CategoryResult, rawResponse json.RawMessage, err error) {
	rawResponse, err = Request(siteConfig, methodGet, uriCategoryInfo, url.Values{
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
