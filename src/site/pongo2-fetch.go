package site

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"sersh.com/totaltube/frontend/internal"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/objx"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

type tagFetchNode struct {
	what    string
	wrapper *pongo2.NodeWrapper
	args    map[string]pongo2.IEvaluator
	headers []pongo2.IEvaluator
	timeout pongo2.IEvaluator
	method  pongo2.IEvaluator
	cache   pongo2.IEvaluator
	raw     pongo2.IEvaluator // get raw response as string, no marshalling to json?
}

var headerRegex = regexp.MustCompile(`^\s*([^:]+)\s*:\s*(.*?)\s*$`)

func (node *tagFetchNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	fetchContext := pongo2.NewChildExecutionContext(ctx)
	var cacheTimeout time.Duration
	if node.cache != nil {
		if cacheTimeoutE, err := node.cache.Evaluate(fetchContext); err != nil {
			return err
		} else {
			cacheTimeout = time.Second * time.Duration(cacheTimeoutE.Integer())
		}
	}
	host, _ := ctx.Public["host"].(string)
	nocache, _ := ctx.Public["nocache"].(bool)
	config := ctx.Public["config"].(*types.Config)
	if strings.HasPrefix(node.what, "http://") || strings.HasPrefix(node.what, "https://") {
		// Fetching information from user address
		// First let's check if we have cache
		cacheKey := ""
		f := helpers.SiteFetch(config)(node.what)
		method := "GET"
		if node.method != nil {
			m, err := node.method.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			if m.String() != "" {
				method = strings.ToUpper(m.String())
				f.WithMethod(method)
			}
		}
		if node.timeout != nil {
			t, err := node.timeout.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			if t.String() != "" {
				timeout := types.ParseHumanDuration(t.String())
				if timeout > 0 {
					f.WithTimeout(int64(timeout / time.Second))
				}
			}
		}
		var isForm = false
		for _, h := range node.headers {
			hh, err := h.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			matches := headerRegex.FindStringSubmatch(hh.String())
			if matches == nil {
				log.Println("wrong header", hh.String(), ". Must be in format [HeaderKey]:[HeaderValue]")
				continue
			}
			if strings.HasPrefix(matches[2], "application/x-www-form-urlencoded") {
				isForm = true
			}
			f.WithHeader(matches[1], matches[2])
		}
		var params = url.Values{}
		if method == "GET" || method == "DELETE" {
			for k, v := range node.args {
				pv, err := v.Evaluate(fetchContext)
				if err != nil {
					return err
				}
				f.WithQueryParam(k, pv.String())
				params.Add(k, pv.String())
			}
		} else {
			data := map[string]interface{}{}
			for k, v := range node.args {
				pv, err := v.Evaluate(fetchContext)
				if err != nil {
					return err
				}
				data[k] = pv.Interface()
				params.Add(k, fmt.Sprintf("%v", pv.Interface()))
			}
			if len(data) > 0 {
				if isForm {
					f.WithFormData(data)
				} else {
					f.WithJsonData(data)
				}
			}
		}
		isRaw := false
		if node.raw != nil {
			raw, err := node.raw.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			isRaw = raw.Bool()
		}
		if cacheTimeout > 0 {
			cacheKey = "in:fetch:" + helpers.Md5Hash(fmt.Sprintf("%s|%s|%s", node.what, method, params.Encode()))
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout, func() ([]byte, error) {
				return f.Do()
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			if isRaw {
				fetchContext.Private["fetch_response"] = string(cached)
			} else {
				parsed, err := objx.FromJSON(string(cached))
				if err != nil {
					log.Println(err)
					return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
				}
				fetchContext.Private["fetch_response"] = parsed
			}
		} else {
			raw, err := f.Do()
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			if isRaw {
				fetchContext.Private["fetch_response"] = string(raw)
			} else {
				parsed, err := objx.FromJSON(string(raw))
				if err != nil {
					log.Println(err)
					return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
				}
				fetchContext.Private["fetch_response"] = parsed
			}
		}
		err := node.wrapper.Execute(fetchContext, writer)
		return err
	}
	amount := 100
	argAmount := 0
	if a, ok := node.args["amount"]; ok {
		av, err := a.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		amount = av.Integer()
		argAmount = av.Integer()
		if amount <= 0 {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("amount must be > 0"),
			}
		}
	}
	page := 1
	if p, ok := node.args["page"]; ok {
		pv, err := p.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		page = pv.Integer()
		if page < 1 {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("page number must be >= 1"),
			}
		}
	}
	lang := "en"
	if l, ok := fetchContext.Public["lang"]; ok {
		lang = l.(*types.Language).Id
	}
	if l, ok := node.args["lang"]; ok {
		lv, err := l.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		lang = lv.String()
		if lang == "" {
			lang = "en"
		}
	}
	sort := api.SortPopular
	argSort := ""
	if s, ok := node.args["sort"]; ok {
		sv, err := s.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		sort = api.SortBy(sv.String())
		argSort = sv.String()
	}
	searchQuery := ""
	if s, ok := node.args["search_query"]; ok {
		sv, err := s.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		searchQuery = sv.String()
	}
	minSearches := 1
	if s, ok := node.args["min_searches"]; ok {
		cv, err := s.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		minSearches = cv.Integer()
	}
	group, _ := ctx.Public["country_group"].(types.CountryGroup)
	cacheKey := ""
	if cacheTimeout > 0 {
		cacheKey = "in:fetch:" + host + ":" + helpers.Md5Hash(fmt.Sprintf("%s|%d|%d|%v|%s|%s|%v|%d|%d", node.what, amount, page, sort,
			searchQuery, lang, cacheTimeout, minSearches, group.Id))
	}
	switch node.what {
	case "content":
		if sort != api.SortPopular && sort != api.SortRand && sort != api.SortDuration &&
			sort != api.SortDated && sort != api.SortViews && sort != api.SortTitle && sort != api.SortRandNoPaging {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("invalid sort param"),
			}
		}
		categorySlug := ""
		categoryId := int64(0)
		if s, ok := node.args["category_slug"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			categorySlug = sv.String()
		}
		if s, ok := node.args["category_id"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			categoryId = int64(sv.Integer())
		}
		channelSlug := ""
		channelId := int64(0)
		if s, ok := node.args["channel_slug"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			channelSlug = sv.String()
		}
		if s, ok := node.args["channel_id"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			channelId = int64(sv.Integer())
		}
		modelSlug := ""
		modelId := int64(0)
		if s, ok := node.args["model_slug"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			modelSlug = sv.String()
		}
		if s, ok := node.args["model_id"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			modelId = int64(sv.Integer())
		}
		timeframe := ""
		if s, ok := node.args["timeframe"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			timeframe = sv.String()
		}
		tag := ""
		if s, ok := node.args["tag"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			tag = sv.String()
		}
		durationGte := int64(0)
		durationLt := int64(0)
		if s, ok := node.args["duration_gte"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			durationGte = int64(sv.Integer())
		}
		if s, ok := node.args["duration_lt"]; ok {
			sv, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			durationLt = int64(sv.Integer())
		}
		userAgent := ""
		if headers, ok := ctx.Public["headers"].(map[string]string); ok {
			userAgent = headers["User-Agent"]
		}
		ip := "127.0.0.1"
		ip, _ = ctx.Public["ip"].(string)
		if cacheTimeout > 0 {
			cacheKey = "in:fetch:" + host + ":" + helpers.Md5Hash(fmt.Sprintf("%s|%d|%d|%v|%s|%s|%v|%d|%s|%d|%s|%d|%s|%s|%s|%d|%d", node.what, amount, page, sort,
				searchQuery, lang, cacheTimeout, categoryId, categorySlug, channelId, channelSlug,
				modelId, modelSlug, timeframe, tag, durationGte, durationLt))
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout/2, func() ([]byte, error) {
				_, rawResponse, err := api.Content(config, api.ContentParams{
					Ip:           net.ParseIP(ip),
					Lang:         lang,
					Page:         int64(page),
					Amount:       int64(amount),
					CategoryId:   categoryId,
					CategorySlug: categorySlug,
					ChannelId:    channelId,
					ChannelSlug:  channelSlug,
					ModelId:      modelId,
					ModelSlug:    modelSlug,
					Sort:         sort,
					Timeframe:    timeframe,
					Tag:          tag,
					DurationGte:  durationGte,
					DurationLt:   durationLt,
					SearchQuery:  searchQuery,
					UserAgent:    userAgent,
					GroupId:      group.Id,
				})
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			var results = new(types.ContentResults)
			if err = json.Unmarshal(cached, results); err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			} else {
				fetchContext.Private["fetched_content"] = results
			}
		} else {
			results, _, err := api.Content(config, api.ContentParams{
				Ip:           net.ParseIP(ip),
				Lang:         lang,
				Page:         int64(page),
				Amount:       int64(amount),
				CategoryId:   categoryId,
				CategorySlug: categorySlug,
				ChannelId:    channelId,
				ChannelSlug:  channelSlug,
				ModelId:      modelId,
				ModelSlug:    modelSlug,
				Sort:         sort,
				Timeframe:    timeframe,
				Tag:          tag,
				DurationGte:  durationGte,
				DurationLt:   durationLt,
				SearchQuery:  searchQuery,
				UserAgent:    userAgent,
				GroupId:      group.Id,
			})
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["fetched_content"] = results
		}
	case "categories":
		if sort != api.SortTitle && sort != api.SortTotal && sort != api.SortPopular {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("invalid sort param"),
			}
		}
		if cacheTimeout > 0 {
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout, func() ([]byte, error) {
				_, rawResponse, err := api.CategoriesList(config, lang, int64(page), sort, int64(amount), group.Id)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			var results = new(types.CategoryResults)
			if err = json.Unmarshal(cached, results); err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			} else {
				fetchContext.Private["categories"] = results
			}
		} else {
			results, _, err := api.CategoriesList(config, lang, int64(page), sort, int64(amount), group.Id)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["categories"] = results
		}
	case "models":
		if sort != api.SortTitle && sort != api.SortTotal && sort != api.SortPopular {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("invalid sort param"),
			}
		}
		if cacheTimeout > 0 {
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout, func() ([]byte, error) {
				_, rawResponse, err := api.ModelsList(config, lang, int64(page), sort, int64(amount), searchQuery, group.Id)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			var results = new(types.ModelResults)
			if err = json.Unmarshal(cached, results); err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			} else {
				fetchContext.Private["models"] = results
			}
		} else {
			results, _, err := api.ModelsList(config, lang, int64(page), sort, int64(amount), searchQuery, group.Id)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["models"] = results
		}
	case "channels":
		if sort != api.SortTitle && sort != api.SortTotal && sort != api.SortPopular {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("invalid sort param"),
			}
		}
		if cacheTimeout > 0 {
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout, func() ([]byte, error) {
				_, rawResponse, err := api.ChannelsList(config, lang, int64(page), sort, int64(amount), group.Id)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			var results = new(types.ChannelResults)
			if err = json.Unmarshal(cached, results); err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			} else {
				fetchContext.Private["channels"] = results
			}
		} else {
			results, _, err := api.ChannelsList(config, lang, int64(page), sort, int64(amount), group.Id)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["channels"] = results
		}
	case "comments":
		var contentId int64
		if node.args["content_id"] != nil {
			contentIdE, err := node.args["content_id"].Evaluate(fetchContext)
			if err != nil {
				return err
			}
			contentId = int64(contentIdE.Integer())
		} else if contentItem, ok := ctx.Public["content_item"].(*types.ContentItemResult); ok {
			contentId = contentItem.Id
		} else {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("content_item not found"),
			}
		}
		from := 0
		size := internal.Config.Comments.ItemsPerPage
		commentsSort := api.CommentsSortByLastUpdated
		if argSort != "" {
			commentsSort = api.CommentsSortBy(argSort)
			if commentsSort != api.CommentsSortByLastUpdated && commentsSort != api.CommentsSortByLikes &&
				commentsSort != api.CommentsSortByDislikes && commentsSort != api.CommentsSortByCreated {
				return &pongo2.Error{
					Sender:    "tag:fetch",
					OrigError: errors.New("invalid sort param"),
				}
			}
		}
		if s, ok := node.args["from"]; ok {
			fromE, err := s.Evaluate(fetchContext)
			if err != nil {
				return err
			}
			from = fromE.Integer()
		}
		if argAmount > 0 {
			size = argAmount
		}
		if cacheTimeout > 0 {
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout, func() ([]byte, error) {
				_, rawResponse, err := api.GetComments(config, contentId, from, size, commentsSort, lang)
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			var results = new(types.CommentsResult)
			if err = json.Unmarshal(cached, results); err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			} else {
				fetchContext.Private["comments"] = results
			}
		} else {
			results, _, err := api.GetComments(config, contentId, from, size, commentsSort, lang)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["comments"] = results
		}
	case "searches":
		if cacheTimeout > 0 {
			cached, err := db.GetCachedTimeout(cacheKey, cacheTimeout, cacheTimeout, func() ([]byte, error) {
				var rawResponse []byte
				var err error
				if sort == api.SortRand {
					_, rawResponse, err = api.RandomSearches(config, lang, int64(amount), int64(minSearches))
				} else {
					_, rawResponse, err = api.TopSearches(config, lang, int64(amount))
				}
				return rawResponse, err
			}, nocache)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			var result struct {
				Items []types.TopSearch `json:"items"`
			}
			if err = json.Unmarshal(cached, &result); err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			} else {
				fetchContext.Private["searches"] = result.Items
			}
		} else {
			var results []types.TopSearch
			var err error
			if sort == api.SortRand {
				results, _, err = api.RandomSearches(config, lang, int64(amount), int64(minSearches))
			} else {
				results, _, err = api.TopSearches(config, lang, int64(amount))
			}
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["searches"] = results
		}
	}
	Err := node.wrapper.Execute(fetchContext, writer)
	return Err
}

func pongo2Fetch(doc *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagFetch := &tagFetchNode{
		headers: make([]pongo2.IEvaluator, 0, 5),
	}
	var err *pongo2.Error
	whatToken := arguments.MatchType(pongo2.TokenString)
	if whatToken == nil {
		return nil, arguments.Error("Expected string - one of: categories, channels, searches, models, content or url", nil)
	}
	tagFetch.what = whatToken.Val
	tagFetch.args = make(map[string]pongo2.IEvaluator)
	for {
		commaToken := arguments.MatchType(pongo2.TokenSymbol)
		if commaToken == nil {
			break
		}
		if commaToken.Val != "," {
			return nil, arguments.Error("Comma symbol expected", commaToken)
		}
		idToken := arguments.MatchType(pongo2.TokenIdentifier)
		if idToken == nil {
			return nil, arguments.Error("Identifier expected", commaToken)
		}
		equalToken := arguments.MatchType(pongo2.TokenSymbol)
		if equalToken == nil || equalToken.Val != "=" {
			return nil, arguments.Error("= expected", idToken)
		}
		expression, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		if idToken.Val == "header" {
			tagFetch.headers = append(tagFetch.headers, expression)
		} else if idToken.Val == "timeout" {
			tagFetch.timeout = expression
		} else if idToken.Val == "method" {
			tagFetch.method = expression
		} else if idToken.Val == "cache" {
			tagFetch.cache = expression
		} else if idToken.Val == "raw" {
			tagFetch.raw = expression
		} else {
			tagFetch.args[strings.TrimPrefix(idToken.Val, "arg_")] = expression
		}
	}
	tagFetch.wrapper, _, err = doc.WrapUntilTag("endfetch")
	if err != nil {
		return nil, err
	}
	return tagFetch, nil
}
