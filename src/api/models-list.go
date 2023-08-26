package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func ModelsList(siteDomain, lang string, page int64, sort SortBy, amount int64, searchQuery string, groupId int64) (
	results *types.ModelResults, response json.RawMessage, err error) {
	response, err = ModelsListRaw(siteDomain, lang, page, sort, amount, searchQuery, groupId)
	if err != nil {
		return
	}
	results = new(types.ModelResults)
	err = json.Unmarshal(response, &results)
	return
}

func ModelsListRaw(siteDomain, lang string, page int64, sort SortBy, amount int64, searchQuery string, groupId int64) (response json.RawMessage, err error) {
	response, err = Request(siteDomain, methodGet, uriModelsList, url.Values{
		"lang":     []string{lang},
		"sort":     []string{string(sort)},
		"amount":   []string{strconv.FormatInt(amount, 10)},
		"page":     []string{strconv.FormatInt(page, 10)},
		"query":    []string{searchQuery},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	return
}