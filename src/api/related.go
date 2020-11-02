package api

import (
	"encoding/json"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func Related(
	lang string, id int64, slug string, message string, Type types.ContentType, amount int64,
) (results []types.RelatedItem, err error) {
	var response json.RawMessage
	var data = url.Values{}
	if lang != "" {
		data.Add("lang", lang)
	}
	if id > 0 {
		data.Add("id", strconv.FormatInt(id, 10))
	}
	if slug != "" {
		data.Add("slug", slug)
	}
	if message != "" {
		data.Add("message", message)
	}
	if Type != "" {
		data.Add("type", string(Type))
	}
	if amount > 0 {
		data.Add("amount", strconv.FormatInt(amount, 10))
	}
	response, err = apiRequest(methodGet, uriRelated, data)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(response, &results)
	return
}
