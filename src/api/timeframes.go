package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/types"
)

func Timeframes() (siteDomain string, results []types.Timeframe, err error) {
	var response json.RawMessage
	response, err = ApiRequest(siteDomain, methodGet, uriTimeframes, url.Values{})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
