package api

import (
	"github.com/segmentio/encoding/json"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func ContentItem(
	siteDomain, lang, slug string, id int64,
	omitRelatedForLink bool, relatedAmount int64,
) (results *types.ContentItemResult, err error) {
	var response json.RawMessage
	response, err = ApiRequest(siteDomain, methodGet, uriContentItem, url.Values{
		"lang":    []string{lang},
		"slug":    []string{slug},
		"id":      []string{strconv.FormatInt(id, 10)},
		"orfl":    []string{strconv.FormatBool(omitRelatedForLink)},
		"related": []string{strconv.FormatInt(relatedAmount, 10)},
	})
	if err != nil {
		return
	}
	results = new(types.ContentItemResult)
	err = json.Unmarshal(response, results)
	if err != nil {
		log.Println(err, string(response))
	}
	return
}
