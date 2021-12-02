package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Autocomplete(siteDomain, query, lang string) (results *types.AutocompleteResults, err error) {
	var response json.RawMessage
	response, err = ApiRequest(siteDomain, methodGet, uriAutocomplete, url.Values{
		"lang":  []string{lang},
		"query": []string{query},
	})
	if err != nil {
		return
	}
	results = new(types.AutocompleteResults)
	err = json.Unmarshal(response, results)
	return
}
