package api

import (
	"encoding/json"
	"net/url"
)

type ContentIdResult struct {
	Id int64 `json:"id"`
}

// ContentIdBySlug получает ID контента по slug через API миньона
func ContentIdBySlug(siteDomain, slug string) (result *ContentIdResult, err error) {
	var response json.RawMessage
	response, err = ContentIdBySlugRaw(siteDomain, slug)
	if err != nil {
		return
	}
	result = new(ContentIdResult)
	err = json.Unmarshal(response, result)
	return
}

// ContentIdBySlugRaw получает сырой ответ от API для получения ID контента по slug
func ContentIdBySlugRaw(siteDomain, slug string) (response json.RawMessage, err error) {
	params := url.Values{}
	params.Add("slug", slug)
	response, err = Request(siteDomain, methodGet, uriContentId, params)
	return
}
