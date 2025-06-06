package api

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/types"
)

func ContentItem(siteDomain, lang, slug string, id int64, omitRelatedForLink bool, relatedAmount int64, groupId int64,
	relatedTitleTranslated *bool, relatedTitleTranslatedMinTermFreq *int, relatedTitleTranslatedMaxQueryTerms *int, relatedTitleTranslatedBoost *float64,
	relatedTitle *bool, relatedTitleMinTermFreq *int, relatedTitleMaxQueryTerms *int, relatedTitleBoost *float64,
	relatedTags *bool, relatedTagsMinTermFreq *int, relatedTagsMaxQueryTerms *int, relatedTagsBoost *float64) (
	results *types.ContentItemResult, err error) {
	var response json.RawMessage
	response, err = ContentItemRaw(siteDomain, lang, slug, id, omitRelatedForLink, relatedAmount, groupId,
		relatedTitleTranslated, relatedTitleTranslatedMinTermFreq, relatedTitleTranslatedMaxQueryTerms, relatedTitleTranslatedBoost,
		relatedTitle, relatedTitleMinTermFreq, relatedTitleMaxQueryTerms, relatedTitleBoost,
		relatedTags, relatedTagsMinTermFreq, relatedTagsMaxQueryTerms, relatedTagsBoost)
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

func ContentItemRaw(siteDomain, lang, slug string, id int64, omitRelatedForLink bool, relatedAmount int64, groupId int64,
	relatedTitleTranslated *bool, relatedTitleTranslatedMinTermFreq *int, relatedTitleTranslatedMaxQueryTerms *int, relatedTitleTranslatedBoost *float64,
	relatedTitle *bool, relatedTitleMinTermFreq *int, relatedTitleMaxQueryTerms *int, relatedTitleBoost *float64,
	relatedTags *bool, relatedTagsMinTermFreq *int, relatedTagsMaxQueryTerms *int, relatedTagsBoost *float64) (response json.RawMessage, err error) {
	params := url.Values{}
	if relatedTitleTranslated != nil {
		params.Add("related_title_translated", strconv.FormatBool(*relatedTitleTranslated))
	}
	if relatedTitleTranslatedMinTermFreq != nil {
		params.Add("related_title_translated_min_term_freq", strconv.Itoa(*relatedTitleTranslatedMinTermFreq))
	}
	if relatedTitleTranslatedMaxQueryTerms != nil {
		params.Add("related_title_translated_max_query_terms", strconv.Itoa(*relatedTitleTranslatedMaxQueryTerms))
	}
	if relatedTitle != nil {
		params.Add("related_title", strconv.FormatBool(*relatedTitle))
	}
	if relatedTitleMinTermFreq != nil {
		params.Add("related_title_min_term_freq", strconv.Itoa(*relatedTitleMinTermFreq))
	}
	if relatedTitleMaxQueryTerms != nil {
		params.Add("related_title_max_query_terms", strconv.Itoa(*relatedTitleMaxQueryTerms))
	}
	if relatedTitleBoost != nil {
		params.Add("related_title_boost", strconv.FormatFloat(*relatedTitleBoost, 'f', -1, 64))
	}
	if relatedTags != nil {
		params.Add("related_tags", strconv.FormatBool(*relatedTags))
	}
	if relatedTagsMinTermFreq != nil {
		params.Add("related_tags_min_term_freq", strconv.Itoa(*relatedTagsMinTermFreq))
	}
	if relatedTagsMaxQueryTerms != nil {
		params.Add("related_tags_max_query_terms", strconv.Itoa(*relatedTagsMaxQueryTerms))
	}
	if relatedTagsBoost != nil {
		params.Add("related_tags_boost", strconv.FormatFloat(*relatedTagsBoost, 'f', -1, 64))
	}
	params.Add("lang", lang)
	params.Add("slug", slug)
	params.Add("id", strconv.FormatInt(id, 10))
	params.Add("orfl", strconv.FormatBool(omitRelatedForLink))
	params.Add("related", strconv.FormatInt(relatedAmount, 10))
	params.Add("group_id", strconv.FormatInt(groupId, 10))
	response, err = Request(siteDomain, methodGet, uriContentItem, params)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Println(err, "slug: ", slug, "id: ", id)
	}
	return
}
