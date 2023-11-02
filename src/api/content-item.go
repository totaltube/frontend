package api

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/types"
)

func ContentItem(siteDomain, lang, slug string, id int64, omitRelatedForLink bool, relatedAmount int64, groupId int64) (
	results *types.ContentItemResult, err error) {
	var response json.RawMessage
	response, err = ContentItemRaw(siteDomain, lang, slug, id, omitRelatedForLink, relatedAmount, groupId)
	if err != nil {
		return
	}
	results = new(types.ContentItemResult)
	err = json.Unmarshal(response, results)
	if err != nil {
		log.Println(err, string(response))
	}
	format := results.GetThumbFormat()
	results.ThumbFormat = format.Name
	results.ThumbsWidth = int32(format.Width)
	results.ThumbsHeight = int32(format.Height)
	results.ThumbsAmount = int32(format.Amount)
	results.ThumbRetina = format.Retina
	results.ThumbType = format.Type
	results.ThumbWidth = results.ThumbsHeight
	results.ThumbHeight = results.ThumbsHeight
	return
}

func ContentItemRaw(siteDomain, lang, slug string, id int64, omitRelatedForLink bool, relatedAmount int64, groupId int64) (response json.RawMessage, err error) {
	response, err = Request(siteDomain, methodGet, uriContentItem, url.Values{
		"lang":     []string{lang},
		"slug":     []string{slug},
		"id":       []string{strconv.FormatInt(id, 10)},
		"orfl":     []string{strconv.FormatBool(omitRelatedForLink)},
		"related":  []string{strconv.FormatInt(relatedAmount, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Println(err, "slug: ", slug, "id: ", id)
	}
	return
}
