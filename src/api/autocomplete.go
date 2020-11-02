package api

import (
	"github.com/segmentio/encoding/json"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Autocomplete(query, lang string) (results *types.AutocompleteResults, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriAutocomplete, url.Values{
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
