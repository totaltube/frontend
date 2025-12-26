package api

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/types"
)

func TopContent(siteConfig *types.Config, lang string, page int64, groupId int64, additionalLanguages []string) (results *types.ContentResults, err error) {
	var response json.RawMessage
	response, err = TopContentRaw(siteConfig, lang, page, groupId, additionalLanguages)
	if err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(response, results)
	return
}

func TopContentRaw(siteConfig *types.Config, lang string, page int64, groupId int64, additionalLanguages []string) (response json.RawMessage, err error) {
	data := url.Values{
		"lang":     []string{lang},
		"page":     []string{strconv.FormatInt(page, 10)},
		"group_id": []string{strconv.FormatInt(groupId, 10)},
	}
	if len(additionalLanguages) > 0 {
		data.Add("additional_languages", strings.Join(additionalLanguages, ","))
	}
	return Request(siteConfig, methodGet, uriTopContent, data)
}
