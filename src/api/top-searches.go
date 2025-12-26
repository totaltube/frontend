package api

import (
	"encoding/json"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func TopSearches(siteConfig *types.Config, lang string, amount int64) (results []types.TopSearch, response json.RawMessage, err error) {
	response, err = Request(siteConfig, methodGet, uriTopSearches, url.Values{
		"lang":   []string{lang},
		"amount": []string{strconv.FormatInt(amount, 10)},
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
