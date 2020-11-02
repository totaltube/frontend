package api

import (
	"encoding/json"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func TopSearches(lang string, amount int64) (results []types.TopSearch, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodGet, uriTopSearches, url.Values{
		"lang":   []string{lang},
		"amount": []string{strconv.FormatInt(amount, 10)},
	})
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
