package api

import (
	"net/url"
	"strconv"

	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/types"
)

func TopCategories(siteDomain, lang string, page int64, groupId int64) (results *types.CategoryResults, err error) {
	var response json.RawMessage
	response, err = ApiRequest(siteDomain, methodGet, uriTopCategories, url.Values{
		"lang": []string{lang},
		"page": []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(response, results)
	return
}
