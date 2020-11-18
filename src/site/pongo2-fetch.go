package site

import (
	"errors"
	"github.com/flosch/pongo2/v4"
	"log"
	"regexp"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
	"strings"
)

type tagFetchNode struct {
	what    string
	wrapper *pongo2.NodeWrapper
	args    map[string]pongo2.IEvaluator
	headers []pongo2.IEvaluator
	timeout pongo2.IEvaluator
	method  pongo2.IEvaluator
	raw     bool // получить ли "сырой" ответ в виде строки, без маршалинга в json?
}

var headerRegex = regexp.MustCompile(`^\s*([^:]+)\s*:\s*(.*?)\s*$`)

func (node *tagFetchNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	fetchContext := pongo2.NewChildExecutionContext(ctx)
	if strings.HasPrefix(node.what, "http://") || strings.HasPrefix(node.what, "https://") {
		// Особый случай - тут мы фетчим информацию с произвольного URL
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
			f.WithHeader(matches[1], matches[2])
		}
		if method == "GET" || method == "DELETE" {
			for k, v := range node.args {
				pv, err := v.Evaluate(fetchContext)
				if err != nil {
					return err
				}
				f.WithQueryParam(k, pv.String())
			}
		} else {
			data := map[string]interface{}{}
			for k, v := range node.args {
				pv, err := v.Evaluate(fetchContext)
				if err != nil {
					return err
				}
				data[k] = pv.Interface()
			}
			if len(data) > 0 {
				f.WithData(data)
			}
		}
		if node.raw {
			data := f.String()
			fetchContext.Private["fetch_response"] = data
		} else {
			data := f.Json()
			fetchContext.Private["fetch_response"] = data
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
		lang = l.(string)
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
	switch node.what {
	case "categories":
		results, err := api.CategoriesList(lang, int64(page), sort, int64(amount))
		if err != nil {
			log.Println(err)
		}
		fetchContext.Private["categories"] = results
	case "models":
		results, err := api.ModelsList(lang, int64(page), sort, int64(amount), searchQuery)
		if err != nil {
			log.Println(err)
		}
		fetchContext.Private["models"] = results
	case "channels":
		results, err := api.ChannelsList(lang, int64(page), sort, int64(amount))
		if err != nil {
			log.Println(err)
		}
		fetchContext.Private["channels"] = results
	case "searches":
		results, err := api.TopSearches(lang, int64(amount))
		if err != nil {
			log.Println(err)
		}
		fetchContext.Private["searches"] = results
	}
	err := node.wrapper.Execute(fetchContext, writer)
	return err
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
