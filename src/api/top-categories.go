package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func TopCategories(lang string, page int64) (results types.CategoryResults, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriTopCategories, url.Values{
		"lang": []string{lang},
		"page": []string{strconv.FormatInt(page, 10)},
	})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
