package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func Category(siteDomain, lang string, categoryId int64, categorySlug string, page int64) (results *types.ContentResults, err error) {
	var response json.RawMessage
	data := url.Values{
		"id":   []string{strconv.FormatInt(categoryId, 10)},
		"slug": []string{categorySlug},
		"lang": []string{lang},
		"page": []string{strconv.FormatInt(page, 10)},
	}
	response, err = ApiRequest(siteDomain, methodGet, uriCategory, data)
	if err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(response, results)
	return
}
