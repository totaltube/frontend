package api

import (
	"encoding/json"
	"net/url"
	
	"sersh.com/totaltube/frontend/types"
)

type ContentIdResult struct {
	Id int64 `json:"id"`
}

// ContentIdBySlug получает ID контента по slug через API миньона
func ContentIdBySlug(siteConfig *types.Config, slug string) (result *ContentIdResult, err error) {
	var response json.RawMessage
	response, err = ContentIdBySlugRaw(siteConfig, slug)
	if err != nil {
		return
	}
	result = new(ContentIdResult)
	err = json.Unmarshal(response, result)
	return
}

// ContentIdBySlugRaw получает сырой ответ от API для получения ID контента по slug
func ContentIdBySlugRaw(siteConfig *types.Config, slug string) (response json.RawMessage, err error) {
	params := url.Values{}
	params.Add("slug", slug)
	response, err = Request(siteConfig, methodGet, uriContentId, params)
	return
}
