package api

import (
	"net/url"

	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/types"
)

func Autocomplete(siteDomain, query, lang string) (results *types.AutocompleteResults, err error) {
	var response json.RawMessage
	response, err = ApiRequest(siteDomain, methodGet, uriAutocomplete, url.Values{
		"lang":  []string{lang},
		"query": []string{query},
	})
	results = new(types.AutocompleteResults)
	err = json.Unmarshal(response, results)
	return
}
