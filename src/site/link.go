package site

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/samber/lo"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

type linkParam struct {
	Type  string
	Value interface{}
}

func fastRemove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func GetLink(route string, config *Config, pageTemplate string, langId string, args ...interface{}) (link string) {
	var params = make([]linkParam, 0, len(args)/2)
	curType := ""
	for _, p := range args {
		if curType == "" {
			curType = fmt.Sprintf("%v", p)
			continue
		}
		params = append(params, linkParam{Type: curType, Value: p})
		curType = ""
	}
	if route == "" {
		// route not provided
		return
	}
	var isCustomRoute = false
	switch route {
	case "top_categories", "top-categories":
		link = config.Routes.TopCategories
	case "top_content", "top-content":
		link = config.Routes.TopContent
	case "autocomplete":
		link = config.Routes.Autocomplete
	case "popular":
		link = config.Routes.Popular
	case "new":
		link = config.Routes.New
	case "long":
		link = config.Routes.Long
	case "models":
		link = config.Routes.Models
	case "out":
		link = config.Routes.Out
	case "search":
		link = config.Routes.Search
		if strings.Contains(link, "{query}") {
			queryParam, queryIndex, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "query" })
			if !ok {
				log.Println("no query param for search route")
			} else {
				link = strings.ReplaceAll(link, "{query}", fmt.Sprintf("%v", queryParam.Value))
				params = fastRemove(params, queryIndex)
			}
		}
	case "fake_player", "fake-player":
		link = config.Routes.FakePlayer
		if slugParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "slug" }); ok {
			link = strings.ReplaceAll(link, "{slug}", fmt.Sprintf("%v", slugParam.Value))
			params = fastRemove(params, index)
		}
		if idParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "id" }); ok {
			link = strings.ReplaceAll(link, "{id}", fmt.Sprintf("%v", idParam.Value))
			params = fastRemove(params, index)
		}
		var categories []types.TaxonomyResult
		if categoriesParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "categories"}); ok {
			if categories, ok = categoriesParam.Value.(types.TaxonomyResults); ok {
				fastRemove(params, index)
			}
		}
		if categoryParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "category" }); ok {
			link = strings.ReplaceAll(link, "{category}", fmt.Sprintf("%v", categoryParam.Value))
			params = fastRemove(params, index)
		} else {
			category := "default"
			if len(categories) > 0 {
				category = categories[0].Slug
			}
			link = strings.ReplaceAll(link, "{category}", category)
		}
	case "model":
		link = config.Routes.Model
		if slugParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "slug" }); ok {
			link = strings.ReplaceAll(link, "{slug}", fmt.Sprintf("%v", slugParam.Value))
			params = fastRemove(params, index)
		}
		if idParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "id" }); ok {
			link = strings.ReplaceAll(link, "{id}", fmt.Sprintf("%v", idParam.Value))
			params = fastRemove(params, index)
		}
	case "category":
		link = config.Routes.Category
		if slugParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "slug" }); ok {
			link = strings.ReplaceAll(link, "{slug}", fmt.Sprintf("%v", slugParam.Value))
			params = fastRemove(params, index)
		}
		if idParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "id" }); ok {
			link = strings.ReplaceAll(link, "{id}", fmt.Sprintf("%v", idParam.Value))
			params = fastRemove(params, index)
		}
	case "channel":
		link = config.Routes.Channel
		if slugParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "slug" }); ok {
			link = strings.ReplaceAll(link, "{slug}", fmt.Sprintf("%v", slugParam.Value))
			params = fastRemove(params, index)
		}
		if idParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "id" }); ok {
			link = strings.ReplaceAll(link, "{id}", fmt.Sprintf("%v", idParam.Value))
			params = fastRemove(params, index)
		}
	case "content", "content-item", "content_item":
		var categories []types.TaxonomyResult
		if categoriesParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "categories"}); ok {
			if categories, ok = categoriesParam.Value.(types.TaxonomyResults); ok {
				fastRemove(params, index)
			}
		}
		link = config.Routes.ContentItem
		if slugParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "slug" }); ok {
			link = strings.ReplaceAll(link, "{slug}", fmt.Sprintf("%v", slugParam.Value))
			params = fastRemove(params, index)
		}
		if idParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "id" }); ok {
			link = strings.ReplaceAll(link, "{id}", fmt.Sprintf("%v", idParam.Value))
			params = fastRemove(params, index)
		}
		if categoryParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "category" }); ok {
			link = strings.ReplaceAll(link, "{category}", fmt.Sprintf("%v", categoryParam.Value))
			params = fastRemove(params, index)
		} else {
			category := "default"
			if len(categories) > 0 {
				category = categories[0].Slug
			}
			link = strings.ReplaceAll(link, "{category}", category)
		}
	default:
		route = strings.TrimPrefix(route, "custom.")
		if r, ok := config.Routes.Custom[route]; ok {
			link = r
			isCustomRoute = true
		} else {
			link = route
		}
		if config.General.MultiLanguage {
			link = strings.ReplaceAll(link, "{lang}", langId)
		}
		link, _ = paramRegex.ReplaceFunc(link, func(match regexp2.Match) string {
			if param, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == match.Groups()[1].String() }); ok {
				params = fastRemove(params, index)
				return url.PathEscape(fmt.Sprintf("%v", param.Value))
			}
			return match.String()
		}, -1, -1)
	}
	if config.General.MultiLanguage && !httpRegex.MatchString(link) && !isCustomRoute {
		link = strings.ReplaceAll(config.Routes.LanguageTemplate, "{route}", link)
		link = strings.ReplaceAll(link, "{lang}", langId)
	}
	var pageNum int64 = 1
	if strings.Contains(link, "{page}") {
		pageParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "page" })
		if ok {
			pageNum, _ = strconv.ParseInt(fmt.Sprintf("%v", pageParam.Value), 10, 64)
			if pageNum < 1 {
				pageNum = 1
			}
			params = fastRemove(params, index)
		}
		link = strings.ReplaceAll(link, "{page}", fmt.Sprintf("%d", pageNum))
	}
	isOut := false
	if outParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "out" }); ok {
		isOut, _ = strconv.ParseBool(fmt.Sprintf("%v", outParam.Value))
		params = fastRemove(params, index)
	}
	if isOut && !lo.Contains([]string{"category", "content"}, route) {
		isOut = false
	}
	withTrade := false
	if withTradeParam, index, ok := lo.FindIndexOf(params, func(p linkParam) bool { return p.Type == "with_trade" }); ok {
		withTrade, _ = strconv.ParseBool(fmt.Sprintf("%v", withTradeParam.Value))
		params = fastRemove(params, index)
	}
	urlParams := url.Values{}
	for _, p := range params {
		key := p.Type
		switch key {
		case "content_id":
			key = config.Params.ContentId
		case "content_slug":
			key = config.Params.ContentSlug
		case "category_slug":
			key = config.Params.CategorySlug
		case "category_id":
			key = config.Params.CategoryId
		case "model_slug":
			key = config.Params.ModelSlug
		case "model_id":
			key = config.Params.ModelId
		case "channel_slug":
			key = config.Params.ChannelSlug
		case "channel_id":
			key = config.Params.ChannelId
		case "duration_gte":
			key = config.Params.DurationGte
		case "duration_lt":
			key = config.Params.DurationLt
		case "search_query":
			key = config.Params.SearchQuery
		case "sort_by":
			key = config.Params.SortBy
		case "sort_by_views":
			key = config.Params.SortByViews
		case "sort_by_views_timeframe":
			key = config.Params.SortByViewsTimeframe
		case "sort_by_duration":
			key = config.Params.SortByDuration
		case "sort_by_date":
			key = config.Params.SortByDate
		case "count_redirect":
			key = config.Params.CountRedirect
		case "count_type":
			key = config.Params.CountType
		case "count_thumb_id":
			key = config.Params.CountThumbId
		case "page":
			key = config.Params.Page
		case "nocache":
			key = config.Params.Nocache
		}
		s := fmt.Sprintf("%v", p.Value)
		if key == config.Params.SortBy {
			switch s {
			case "views":
				s = config.Params.SortByViews
			case "duration":
				s = config.Params.SortByDuration
			case "dated":
				s = config.Params.SortByDate
			case "rand":
				s = config.Params.SortByRand
			}
		}
		if key == config.Params.CountType && link == config.Routes.Out {
			switch s {
			case "category":
				s = config.Params.CountTypeCategory
			case "top-categories":
				s = config.Params.CountTypeTopCategories
			case "top-content":
				s = config.Params.CountTypeTopContent
			}
		}
		if s == "" {
			urlParams.Del(key)
		} else {
			urlParams.Set(key, s)
		}
	}
	if isOut {
		// Link to out
		outLink := config.Routes.Out
		outlinkParams := url.Values{}
		countType := config.Params.CountTypeTopCategories
		var categoryId string
		var contentId string
		if urlParams.Has(config.Params.CountType) {
			countType = urlParams.Get(config.Params.CountType)
			urlParams.Del(config.Params.CountType)
		}
		if urlParams.Has(config.Params.CategoryId) {
			categoryId = urlParams.Get(config.Params.CategoryId)
			urlParams.Del(config.Params.CategoryId)
		}
		if urlParams.Has(config.Params.ContentId) {
			contentId = urlParams.Get(config.Params.ContentId)
			urlParams.Del(config.Params.ContentId)
		}
		if len(urlParams) > 0 {
			link = link + "?" + urlParams.Encode()
		}
		outlinkParams.Set(config.Params.CountRedirect, helpers.EncryptBase64(link))
		if countType == config.Params.CountTypeTopCategories || (categoryId != "" && contentId == "") {
			outlinkParams.Set(config.Params.CountType, config.Params.CountTypeTopCategories)
			outlinkParams.Set(config.Params.CategoryId, categoryId)
		} else if countType == config.Params.CountTypeTopContent || (categoryId == "" && contentId != "") {
			outlinkParams.Set(config.Params.CountType, config.Params.CountTypeTopContent)
			outlinkParams.Set(config.Params.ContentId, contentId)
		} else if countType == config.Params.CountTypeCategory || (categoryId != "" && contentId != "") {
			outlinkParams.Set(config.Params.CategoryId, categoryId)
			outlinkParams.Set(config.Params.ContentId, contentId)
			outlinkParams.Set(config.Params.CountType, config.Params.CountTypeCategory)
		}
		if withTrade {
			if strings.Contains(config.General.TradeUrlTemplate, "{{url}}") {
				outlinkParams.Set(config.Params.CountRedirect, helpers.EncryptBase64(strings.ReplaceAll(config.General.TradeUrlTemplate, "{{url}}", link)))
			} else if strings.Contains(config.General.TradeUrlTemplate, "{{encoded_url}}") {
				outlinkParams.Set(config.Params.CountRedirect, helpers.EncryptBase64(strings.ReplaceAll(config.General.TradeUrlTemplate, "{{encoded_url}}", url.QueryEscape(link))))
			} else {
				outlinkParams.Set(config.Params.CountRedirect, config.General.TradeUrlTemplate)
			}
		}
		link = outLink + "?" + outlinkParams.Encode()
		return
	}
	if withTrade {
		if strings.Contains(config.General.TradeUrlTemplate, "{{url}}") {
			link = strings.ReplaceAll(config.General.TradeUrlTemplate, "{{url}}", link)
		} else if strings.Contains(config.General.TradeUrlTemplate, "{{encoded_url}}") {
			link = strings.ReplaceAll(config.General.TradeUrlTemplate, "{{encoded_url}}", url.QueryEscape(link))
		} else {
			link = config.General.TradeUrlTemplate
		}
	}
	return
}
