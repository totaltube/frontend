package api

import (
	"encoding/json"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func RandomSearches(siteDomain, lang string, amount int64, minSearches int64) (results []types.TopSearch, response json.RawMessage, err error) {
	response, err = apiRequest(siteDomain, methodGet, uriRandomSearches, url.Values{
		"lang":         []string{lang},
		"amount":       []string{strconv.FormatInt(amount, 10)},
		"min_searches": []string{strconv.FormatInt(minSearches, 10)},
	})
	if err != nil {
		log.Println(err)
		return
	}
	result := struct {
		Items []types.TopSearch `json:"items"`
	}{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	results = result.Items
	return
}
