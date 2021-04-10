package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func TopContent(siteDomain, lang string, page int64) (results *types.ContentResults, err error) {
	var response json.RawMessage
	response, err = apiRequest(siteDomain, methodGet, uriTopContent, url.Values{
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
