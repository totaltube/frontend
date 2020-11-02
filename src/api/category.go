package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func Category(lang string, categoryId int64, categorySlug string, page int64) (results *types.ContentResults, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriCategory, url.Values{
		"id":   []string{strconv.FormatInt(categoryId, 10)},
		"slug": []string{categorySlug},
		"lang": []string{lang},
		"page": []string{strconv.FormatInt(page, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(response, results)
	return
}
