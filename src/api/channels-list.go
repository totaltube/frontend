package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func ChannelsList(
	lang string, page int64, sort SortBy, amount int64,
) (results *types.ChannelResults, response json.RawMessage, err error) {
	response, err = apiRequest(methodGet, uriChannelsList, url.Values{
		"lang":   []string{lang},
		"sort":   []string{string(sort)},
		"amount": []string{strconv.FormatInt(amount, 10)},
		"page":   []string{strconv.FormatInt(page, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.ChannelResults)
	err = json.Unmarshal(response, &results)
	return
}
