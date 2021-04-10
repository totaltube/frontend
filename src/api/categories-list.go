package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func CategoriesList(
	siteDomain, lang string, page int64, sort SortBy, amount int64,
) (results *types.CategoryResults, rawResponse json.RawMessage, err error) {
	rawResponse, err = apiRequest(siteDomain, methodGet, uriCategoriesList, url.Values{
		"lang":   []string{lang},
		"sort":   []string{string(sort)},
		"amount": []string{strconv.FormatInt(amount, 10)},
		"page":   []string{strconv.FormatInt(page, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.CategoryResults)
	err = json.Unmarshal(rawResponse, &results)
	return
}
