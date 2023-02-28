package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/internal"
)

func Options(siteDomain string) (results *internal.Options, err error) {
	var response json.RawMessage
	response, err = ApiRequest(siteDomain, methodGet, uriOptions, url.Values{})
	if err != nil {
		return
	}
	results = new(internal.Options)
	err = json.Unmarshal(response, results)
	return
}
