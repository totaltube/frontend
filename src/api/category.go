package api

import (
	"net/url"
	"strconv"

	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/types"
)

func Category(siteDomain, lang string, categoryId int64, categorySlug string, page int64, groupId int64) (results *types.ContentResults, err error) {
	var response json.RawMessage
	data := url.Values{
		"id":       []string{strconv.FormatInt(categoryId, 10)},
		"slug":     []string{categorySlug},
		"lang":     []string{lang},
		"page":     []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	}
	response, err = ApiRequest(siteDomain, methodGet, uriCategory, data)
	if err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(response, results)
	return
}
