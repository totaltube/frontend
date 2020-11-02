package api

import (
	"github.com/segmentio/encoding/json"
	"net"
	"net/url"
	"sersh.com/totaltube/frontend/types"
)

func Autocomplete(query, lang string, ip net.IP) (results types.AutocompleteResults, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriTopCategories, url.Values{
		"lang":  []string{lang},
		"query": []string{query},
		"ip":    []string{ip.String()},
	})
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
