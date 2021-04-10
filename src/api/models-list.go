package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func ModelsList(
	siteDomain, lang string, page int64, sort SortBy, amount int64, searchQuery string,
) (results *types.ModelResults, response json.RawMessage, err error) {
	response, err = apiRequest(siteDomain, methodGet, uriModelsList, url.Values{
		"lang":   []string{lang},
		"sort":   []string{string(sort)},
		"amount": []string{strconv.FormatInt(amount, 10)},
		"page":   []string{strconv.FormatInt(page, 10)},
		"query":  []string{searchQuery},
	})
	if err != nil {
		return
	}
	results = new(types.ModelResults)
	err = json.Unmarshal(response, &results)
	return
}
