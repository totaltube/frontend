package site

import (
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/objx"
	"log"
	"net/url"
	"regexp"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
	"strings"
	"time"
)

type tagFetchNode struct {
	what    string
	wrapper *pongo2.NodeWrapper
	args    map[string]pongo2.IEvaluator
	headers []pongo2.IEvaluator
	timeout pongo2.IEvaluator
	method  pongo2.IEvaluator
	cache   pongo2.IEvaluator
	raw     bool // получить ли "сырой" ответ в виде строки, без маршалинга в json?
}

var headerRegex = regexp.MustCompile(`^\s*([^:]+)\s*:\s*(.*?)\s*$`)

func (node *tagFetchNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	fetchContext := pongo2.NewChildExecutionContext(ctx)
	var cacheTimeout time.Duration
	if cacheTimeoutE, err := node.cache.Evaluate(fetchContext); err != nil {
		return err
	} else {
		cacheTimeout = time.Second*time.Duration(cacheTimeoutE.Integer())
	}
	nocache, _ := ctx.Public["nocache"].(bool)
	if strings.HasPrefix(node.what, "http://") || strings.HasPrefix(node.what, "https://") {
		// Fetching information from user address
		// First let's check if we have cache
		cacheKey := ""
		f := helpers.Fetch(node.what)
		m, err := node.method.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		method := "GET"
		if m.String() != "" {
			method = strings.ToUpper(m.String())
			f.WithMethod(method)
		}
		t, err := node.timeout.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		if t.String() != "" {
			timeout := types.ParseHumanDuration(t.String())
			if timeout > 0 {
				f.WithTimeout(timeout)
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
			if matches[2] == "application/x-www-form-urlencoded;charset=UTF-8" {
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
		if cacheTimeout > 0 {
			cacheKey = "fetch:"+helpers.Md5Hash(fmt.Sprintf("%s|%s|%s", node.what, method, params.Encode()))
		}
		if cacheTimeout >0 && !nocache {
			v := db.GetCached(cacheKey)
			if v != nil {
				data := objx.MustFromJSON(string(v))
				fetchContext.Private["fetch_response"] = data
				err = node.wrapper.Execute(fetchContext, writer)
				return err
			}
		}
		if node.raw {
			data := f.String()
			fetchContext.Private["fetch_response"] = data
			if cacheTimeout > 0 {
				err := db.PutCached(cacheKey, []byte(data), cacheTimeout)
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			data := f.Json()
			if data == nil {
				return &pongo2.Error{Sender: "tag:fetch", OrigError: errors.New("can't fetch "+node.what)}
			}
			fetchContext.Private["fetch_response"] = data
			if cacheTimeout > 0 {
				bt, err := json.Marshal(data)
				if err != nil {
					log.Println(err)
				} else {
					err = db.PutCached(cacheKey, bt, cacheTimeout)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
		err = node.wrapper.Execute(fetchContext, writer)
		return err
	}
	amount := 100
	if a, ok := node.args["amount"]; ok {
		av, err := a.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		amount = av.Integer()
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
	if s, ok := node.args["sort"]; ok {
		sv, err := s.Evaluate(fetchContext)
		if err != nil {
			return err
		}
		sort = api.SortBy(sv.String())
		if sort != api.SortPopular && sort != api.SortRand && sort != api.SortDuration &&
			sort != api.SortDated && sort != api.SortViews && sort != api.SortTitle {
			return &pongo2.Error{
				Sender:    "tag:fetch",
				OrigError: errors.New("invalid sort param"),
			}
		}
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
	cacheKey := ""
	if cacheTimeout > 0 {
		cacheKey = "fetch:" + helpers.Md5Hash(fmt.Sprintf("%s|%d|%d|%v|%s|%s|%v|%d", node.what, amount, page, sort,
			searchQuery, lang, cacheTimeout, minSearches))
	}
	var fromCache = false
	switch node.what {
	case "categories":
		if cacheTimeout > 0 && !nocache {
			cached := db.GetCached(cacheKey)
			if cached != nil {
				var results = new(types.CategoryResults)
				if err := json.Unmarshal(cached, results); err != nil {
					log.Println(err)
				} else {
					fetchContext.Private["categories"] = results
					fromCache = true
				}
			}
		}
		if !fromCache {
			results, rawResponse, err := api.CategoriesList(lang, int64(page), sort, int64(amount))
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["categories"] = results
			if cacheTimeout > 0 {
				err = db.PutCached(cacheKey, rawResponse, cacheTimeout)
				if err != nil {
					log.Println(err)
				}
			}
		}
	case "models":
		if cacheTimeout > 0 && !nocache {
			cached := db.GetCached(cacheKey)
			if cached != nil {
				var results = new(types.ModelResults)
				if err := json.Unmarshal(cached, results); err != nil {
					log.Println(err)
				} else {
					fetchContext.Private["models"] = results
					fromCache = true
				}
			}
		}
		if !fromCache {
			results, rawResponse, err := api.ModelsList(lang, int64(page), sort, int64(amount), searchQuery)
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["models"] = results
			if cacheTimeout > 0 {
				err = db.PutCached(cacheKey, rawResponse, cacheTimeout)
				if err != nil {
					log.Println(err)
				}
			}
		}
	case "channels":
		if cacheTimeout > 0 && !nocache {
			cached := db.GetCached(cacheKey)
			if cached != nil {
				var results = new(types.ChannelResults)
				if err := json.Unmarshal(cached, results); err != nil {
					log.Println(err)
				} else {
					fetchContext.Private["channels"] = results
					fromCache = true
				}
			}
		}
		if !fromCache {
			results, rawResponse, err := api.ChannelsList(lang, int64(page), sort, int64(amount))
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["channels"] = results
			if cacheTimeout > 0 {
				err = db.PutCached(cacheKey, rawResponse, cacheTimeout)
				if err != nil {
					log.Println(err)
				}
			}
		}
	case "searches":
		if cacheTimeout > 0 && !nocache {
			cached := db.GetCached(cacheKey)
			if cached != nil {
				var result struct{
					Items []types.TopSearch `json:"items"`
				}
				if err := json.Unmarshal(cached, &result); err != nil {
					log.Println(err)
				} else {
					fetchContext.Private["searches"] = result.Items
					fromCache = true
				}
			}
		}
		if !fromCache {
			var results []types.TopSearch
			var rawResponse []byte
			var err error
			if sort == api.SortRand {
				results, rawResponse, err = api.RandomSearches(lang, int64(amount), int64(minSearches))
			} else {
				results, rawResponse, err = api.TopSearches(lang, int64(amount))
			}
			if err != nil {
				log.Println(err)
				return &pongo2.Error{Sender: "tag:fetch", OrigError: err}
			}
			fetchContext.Private["searches"] = results
			if cacheTimeout > 0 {
				err = db.PutCached(cacheKey, rawResponse, cacheTimeout)
			}
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
		return nil, arguments.Error("Expected string - one of: categories, channels, searches, models or url", nil)
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
