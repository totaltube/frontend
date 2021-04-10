package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Languages(siteDomain string) ( results []types.Language, err error) {
	var response json.RawMessage
	response, err = apiRequest(siteDomain, methodGet, uriLanguages, url.Values{})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
