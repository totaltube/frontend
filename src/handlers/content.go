package handlers

import (
	"fmt"
	"log"
	"net"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func getContentFunc(hostName string, langId string, userAgent string, ip string) func(args ...interface{}) *types.ContentResults {
	return func(args ...interface{}) *types.ContentResults {
		parsingName := true
		params := api.ContentParams{
			Ip:        net.ParseIP(ip),
			Lang:      langId,
			UserAgent: userAgent,
		}
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
				params.Lang = val
			case "page":
				params.Page, _ = strconv.ParseInt(val, 10, 16)
			case "amount":
				params.Amount, _ = strconv.ParseInt(val, 10, 32)
			case "category_id":
				params.CategoryId, _ = strconv.ParseInt(val, 10, 32)
			case "category_slug":
				params.CategorySlug = val
			case "channel_id":
				params.ChannelId, _ = strconv.ParseInt(val, 10, 32)
			case "model_id":
				params.ModelId, _ = strconv.ParseInt(val, 10, 32)
			case "channel_slug":
				params.ChannelSlug = val
			case "model_slug":
				params.ModelSlug = val
			case "related_message":
				params.RelatedMessage = val
			case "sort":
				params.Sort = api.SortBy(val)
			case "timeframe":
				params.Timeframe = val
			case "tag":
				params.Tag = val
			case "duration_gte":
				params.DurationGte, _ = strconv.ParseInt(val, 10, 64)
			case "duration_lt":
				params.DurationLt, _ = strconv.ParseInt(val, 10, 64)
			case "search_query":
				params.SearchQuery = val
			case "is_natural":
				params.IsNatural, _ = strconv.ParseBool(val)
			}
		}
		if results, _, err := api.Content(hostName, params); err != nil {
			log.Println("error getting content: ", err)
			return nil
		} else {
			return results
		}
	}
}
