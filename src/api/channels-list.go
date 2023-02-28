package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func ChannelsList(
	siteDomain, lang string, page int64, sort SortBy, amount int64, groupId int64,
) (results *types.ChannelResults, response json.RawMessage, err error) {
	response, err = ApiRequest(siteDomain, methodGet, uriChannelsList, url.Values{
		"lang":     []string{lang},
		"sort":     []string{string(sort)},
		"amount":   []string{strconv.FormatInt(amount, 10)},
		"page":     []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.ChannelResults)
	err = json.Unmarshal(response, &results)
	return
}
