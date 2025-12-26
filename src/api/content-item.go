package api

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/types"
)

type RelatedParams struct {
	TitleTranslated              *bool
	TitleTranslatedMinTermFreq   *int
	TitleTranslatedMaxQueryTerms *int
	TitleTranslatedBoost         *float64
	RandomizeLast                int
	Title                        *bool
	TitleMinTermFreq             *int
	TitleMaxQueryTerms           *int
	TitleBoost                   *float64
	Tags                         *bool
	TagsMinTermFreq              *int
	TagsMaxQueryTerms            *int
	TagsBoost                    *float64
}

func ContentItem(siteConfig *types.Config, lang, slug string, id int64, omitRelatedForLink bool, relatedAmount int64, groupId int64,
	related *RelatedParams) (results *types.ContentItemResult, err error) {
	var response json.RawMessage
	response, err = ContentItemRaw(siteConfig, lang, slug, id, omitRelatedForLink, relatedAmount, groupId, related)
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

func ContentItemRaw(siteConfig *types.Config, lang, slug string, id int64, omitRelatedForLink bool, relatedAmount int64, groupId int64,
	related *RelatedParams) (response json.RawMessage, err error) {
	params := url.Values{}
	if related != nil {
		if related.RandomizeLast > 0 {
			params.Add("related_randomize_last", strconv.Itoa(related.RandomizeLast))
		}
		if related.TitleTranslated != nil {
			params.Add("related_title_translated", strconv.FormatBool(*related.TitleTranslated))
		}
		if related.TitleTranslatedMinTermFreq != nil {
			params.Add("related_title_translated_min_term_freq", strconv.Itoa(*related.TitleTranslatedMinTermFreq))
		}
		if related.TitleTranslatedMaxQueryTerms != nil {
			params.Add("related_title_translated_max_query_terms", strconv.Itoa(*related.TitleTranslatedMaxQueryTerms))
		}
		if related.TitleTranslatedBoost != nil {
			params.Add("related_title_translated_boost", strconv.FormatFloat(*related.TitleTranslatedBoost, 'f', -1, 64))
		}
		if related.Title != nil {
			params.Add("related_title", strconv.FormatBool(*related.Title))
		}
		if related.TitleMinTermFreq != nil {
			params.Add("related_title_min_term_freq", strconv.Itoa(*related.TitleMinTermFreq))
		}
		if related.TitleMaxQueryTerms != nil {
			params.Add("related_title_max_query_terms", strconv.Itoa(*related.TitleMaxQueryTerms))
		}
		if related.TitleBoost != nil {
			params.Add("related_title_boost", strconv.FormatFloat(*related.TitleBoost, 'f', -1, 64))
		}
		if related.Tags != nil {
			params.Add("related_tags", strconv.FormatBool(*related.Tags))
		}
		if related.TagsMinTermFreq != nil {
			params.Add("related_tags_min_term_freq", strconv.Itoa(*related.TagsMinTermFreq))
		}
		if related.TagsMaxQueryTerms != nil {
			params.Add("related_tags_max_query_terms", strconv.Itoa(*related.TagsMaxQueryTerms))
		}
		if related.TagsBoost != nil {
			params.Add("related_tags_boost", strconv.FormatFloat(*related.TagsBoost, 'f', -1, 64))
		}
	}
	params.Add("lang", lang)
	params.Add("slug", slug)
	params.Add("id", strconv.FormatInt(id, 10))
	params.Add("orfl", strconv.FormatBool(omitRelatedForLink))
	params.Add("related", strconv.FormatInt(relatedAmount, 10))
	params.Add("group_id", strconv.FormatInt(groupId, 10))
	response, err = Request(siteConfig, methodGet, uriContentItem, params)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Println(err, "slug: ", slug, "id: ", id)
	}
	return
}
