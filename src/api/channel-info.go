package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func ChannelInfo(siteDomain, lang string, channelInfo int64, channelSlug string) (result *types.ChannelResult, rawResponse json.RawMessage, err error) {
	rawResponse, err = ApiRequest(siteDomain, methodGet, uriChannelInfo, url.Values{
		"id":   []string{strconv.FormatInt(channelInfo, 10)},
		"slug": []string{channelSlug},
		"lang": []string{lang},
	})
	if err != nil {
		return
	}
	result = new(types.ChannelResult)
	err = json.Unmarshal(rawResponse, result)
	return
}
