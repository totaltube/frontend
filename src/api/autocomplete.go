package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/types"
)

func Autocomplete(siteDomain, query, lang string, config *types.Config) (results *types.AutocompleteResults, err error) {
	var response json.RawMessage
	response, err = Request(siteDomain, methodGet, uriAutocomplete, url.Values{
		"lang":  []string{lang},
		"query": []string{query},
	})
	if err != nil {
		return
	}
	results = new(types.AutocompleteResults)
	err = json.Unmarshal(response, results)
	if err != nil {
		return
	}
	if config.Routes.IdXorKey > 0 {
		for i := range results.Items {
			results.Items[i].Id ^= config.Routes.IdXorKey
		}
	}
	return
}
