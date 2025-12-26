package api

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/types"
)

func Category(siteConfig *types.Config, lang string, categoryId int64, categorySlug string, page int64, groupId int64, additionalLanguages []string) (results *types.ContentResults, err error) {
	var response json.RawMessage
	data := url.Values{
		"id":       []string{strconv.FormatInt(categoryId, 10)},
		"slug":     []string{categorySlug},
		"lang":     []string{lang},
		"page":     []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	}
	if len(additionalLanguages) > 0 {
		data.Add("additional_languages", strings.Join(additionalLanguages, ","))
	}
	response, err = Request(siteConfig, methodGet, uriCategory, data)
	if err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(response, results)
	return
}

func CategoryRaw(siteConfig *types.Config, lang string, categoryId int64, categorySlug string, page int64, groupId int64, additionalLanguages []string) (response json.RawMessage, err error) {
	data := url.Values{
		"id":       []string{strconv.FormatInt(categoryId, 10)},
		"slug":     []string{categorySlug},
		"lang":     []string{lang},
		"page":     []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	}
	if len(additionalLanguages) > 0 {
		data.Add("additional_languages", strings.Join(additionalLanguages, ","))
	}
	response, err = Request(siteConfig, methodGet, uriCategory, data)
	return
}
