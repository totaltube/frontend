package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Options() (siteDomain string, results *types.Options, err error) {
	var response json.RawMessage
	response, err = apiRequest(siteDomain, methodGet, uriOptions, url.Values{})
	if err != nil {
		return
	}
	results = new(types.Options)
	err = json.Unmarshal(response, results)
	return
}
