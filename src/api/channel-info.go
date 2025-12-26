package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func ChannelInfo(siteConfig *types.Config, lang string, channelInfo int64, channelSlug string) (result *types.ChannelResult, rawResponse json.RawMessage, err error) {
	rawResponse, err = Request(siteConfig, methodGet, uriChannelInfo, url.Values{
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
