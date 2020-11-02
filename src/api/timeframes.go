package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Timeframes() (results []types.Timeframe, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriTimeframes, url.Values{})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
