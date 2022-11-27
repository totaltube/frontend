package api

import (
	"net/url"
	"strconv"

	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/types"
)

func CategoriesList(
	siteDomain, lang string, page int64, sort SortBy, amount int64, groupId int64,
) (results *types.CategoryResults, rawResponse json.RawMessage, err error) {
	rawResponse, err = ApiRequest(siteDomain, methodGet, uriCategoriesList, url.Values{
		"lang":     []string{lang},
		"sort":     []string{string(sort)},
		"amount":   []string{strconv.FormatInt(amount, 10)},
		"page":     []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(rawResponse, &results)
	return
}
