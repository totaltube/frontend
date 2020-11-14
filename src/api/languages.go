package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Languages() (results []types.Language, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriLanguages, url.Values{})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
