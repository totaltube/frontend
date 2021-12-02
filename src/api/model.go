package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func ModelInfo(
	siteDomain, lang string, id int64, slug string,
) (results *types.ModelResult, rawResponse json.RawMessage, err error) {
	rawResponse, err = ApiRequest(siteDomain, methodGet, uriModel, url.Values{
		"lang": []string{lang},
		"slug": []string{slug},
		"id":   []string{strconv.FormatInt(id, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.ModelResult)
	err = json.Unmarshal(rawResponse, results)
	return
}
