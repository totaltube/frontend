package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

func Options(siteConfig *types.Config) (results *internal.Options, err error) {
	var response json.RawMessage
	response, err = Request(siteConfig, methodGet, uriOptions, url.Values{})
	if err != nil {
		return
	}
	results = new(internal.Options)
	err = json.Unmarshal(response, results)
	return
}
