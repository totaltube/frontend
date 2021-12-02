package handlers

import (
	"fmt"
	"log"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/types"
	"strconv"
)

func getCategoriesListFunc(hostName string, langId string, defaultAmount int64) func(args ...interface{}) *types.CategoryResults {
	return func(args ...interface{}) *types.CategoryResults {
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
		results, _, err := api.CategoriesList(hostName, langId, page, sortBy, amount)
		if err != nil {
			log.Println("can't get categories list:", err)
			return nil
		}
		return results
	}
}
