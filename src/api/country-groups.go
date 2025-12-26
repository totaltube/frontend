package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/types"
)

func CountryGroups() (results []types.CountryGroup, err error) {
	var response json.RawMessage
	response, err = Request(nil, methodGet, uriCountryGroups, url.Values{})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}