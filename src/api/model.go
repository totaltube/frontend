package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"sersh.com/totaltube/frontend/types"
)

func ModelInfo(
	siteConfig *types.Config, lang string, id int64, slug string, groupId int64,
) (results *types.ModelResult, rawResponse json.RawMessage, err error) {
	rawResponse, err = Request(siteConfig, methodGet, uriModel, url.Values{
		"lang": []string{lang},
		"slug": []string{slug},
		"id":   []string{strconv.FormatInt(id, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.ModelResult)
	err = json.Unmarshal(rawResponse, results)
	return
}
