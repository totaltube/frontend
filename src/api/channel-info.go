package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func ChannelInfo(lang string, channelInfo int64, channelSlug string) (result *types.ChannelResult, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriChannelInfo, url.Values{
		"id":   []string{strconv.FormatInt(channelInfo, 10)},
		"slug": []string{channelSlug},
		"lang": []string{lang},
	})
	if err != nil {
		return
	}
	result = new(types.ChannelResult)
	err = json.Unmarshal(response, result)
	return
}
