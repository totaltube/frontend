package api

import (
	"encoding/json"

	"github.com/tidwall/gjson"
	"sersh.com/totaltube/frontend/types"
)

func BadbotRegister(
	siteConfig *types.Config, ip string, userAgent string, referer string,
) (err error) {
	_, err = Request(siteConfig, methodPost, uriBadbotRegister, M{
		"ip":         ip,
		"user_agent": userAgent,
		"referer":    referer,
	})
	return
}

func GetBadBots() (badBots []string, err error) {
	var response json.RawMessage
	response, err = Request(nil, methodGet, uriBadBotsList, nil)
	if err != nil {
		return
	}
	items := gjson.ParseBytes(response).Get("items").Array()
	badBots = make([]string, 0, len(items))
	for _, item := range items {
		badBots = append(badBots, item.String())
	}
	return
}

func GetWhitelistBots() (whitelistBots []string, err error) {
	var response json.RawMessage
	response, err = Request(nil, methodGet, uriWhitelistBotsList, nil)
	if err != nil {
		return
	}
	items := gjson.ParseBytes(response).Get("items").Array()
	whitelistBots = make([]string, 0, len(items))
	for _, item := range items {
		whitelistBots = append(whitelistBots, item.String())
	}
	return
}
