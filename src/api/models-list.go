package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func ModelsList(siteConfig *types.Config, lang string, page int64, sort SortBy, amount int64, searchQuery string, groupId int64) (
	results *types.ModelResults, response json.RawMessage, err error) {
	response, err = ModelsListRaw(siteConfig, lang, page, sort, amount, searchQuery, groupId)
	if err != nil {
		return
	}
	results = new(types.ModelResults)
	err = json.Unmarshal(response, &results)
	return
}

func ModelsListRaw(siteConfig *types.Config, lang string, page int64, sort SortBy, amount int64, searchQuery string, groupId int64) (response json.RawMessage, err error) {
	response, err = Request(siteConfig, methodGet, uriModelsList, url.Values{
		"lang":     []string{lang},
		"sort":     []string{string(sort)},
		"amount":   []string{strconv.FormatInt(amount, 10)},
		"page":     []string{strconv.FormatInt(page, 10)},
		"query":    []string{searchQuery},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	return
}