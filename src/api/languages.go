package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/types"
)

func Languages(siteConfig *types.Config) (results []types.Language, err error) {
	var response json.RawMessage
	response, err = Request(siteConfig, methodGet, uriLanguages, url.Values{})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
