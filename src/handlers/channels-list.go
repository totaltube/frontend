package handlers

import (
	"fmt"
	"log"
	"strconv"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/types"
)

func getChannelsListFunc(hostName string, langId string, defaultAmount int64, groupId int64) func(args ...interface{}) *types.ChannelResults {
	return func(args ...interface{}) *types.ChannelResults {
		parsingName := true
		var amount = defaultAmount
		var page int64
		var sortBy = api.SortTotal
		curName := ""
		for k := range args {
			if parsingName {
				curName = fmt.Sprintf("%v", args[k])
				parsingName = false
				continue
			}
			val := fmt.Sprintf("%v", args[k])
			parsingName = true
			switch curName {
			case "lang":
				langId = val
			case "page":
				page, _ = strconv.ParseInt(val, 10, 64)
			case "sort":
				sortBy = api.SortBy(val)
			case "amount":
				amount, _ = strconv.ParseInt(val, 10, 64)
			}
		}
		results, _, err := api.ChannelsList(hostName, langId, page, sortBy, amount, groupId)
		if err != nil {
			log.Println("can't get channels list:", err)
			return nil
		}
		return results
	}
}
