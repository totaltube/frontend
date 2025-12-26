package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func TopCategories(siteConfig *types.Config, lang string, page int64, groupId int64) (results *types.CategoryResults, err error) {
	var response json.RawMessage
	response, err = TopCategoriesRaw(siteConfig, lang, page, groupId)
	if err != nil {
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(response, results)
	return
}

func TopCategoriesRaw(siteConfig *types.Config, lang string, page int64, groupId int64) (response json.RawMessage, err error) {
	return Request(siteConfig, methodGet, uriTopCategories, url.Values{
		"lang": []string{lang},
		"page": []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
}