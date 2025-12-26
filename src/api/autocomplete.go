package api

import (
	"encoding/json"
	"net/url"

	"sersh.com/totaltube/frontend/types"
)

func Autocomplete(siteConfig *types.Config, query, lang string) (results *types.AutocompleteResults, err error) {
	var response json.RawMessage
	response, err = Request(siteConfig, methodGet, uriAutocomplete, url.Values{
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
	if siteConfig != nil && siteConfig.Routes.IdXorKey > 0 {
		for i := range results.Items {
			results.Items[i].Id ^= siteConfig.Routes.IdXorKey
		}
	}
	return
}
