package site

import (
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"html/template"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
)

var linksTokens = []string{
	"top_categories", "top_content", "autocomplete", "search", "popular", "new",
	"long", "model", "models", "category", "channel", "content", "out",
}

type tagLinkNode struct {
	what string
	args map[string]pongo2.IEvaluator
}

func (node *tagLinkNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	linkContext := pongo2.NewChildExecutionContext(ctx)
	lang := "en"
	if l, ok := linkContext.Public["lang"]; ok {
		lang = l.(*types.Language).Id
	}
	var copyArgs = make(map[string]pongo2.IEvaluator)
	for k, v := range node.args {
		copyArgs[k] = v
	}
	var config *Config
	if configI, ok := linkContext.Public["config"]; !ok {
		return &pongo2.Error{
			Sender:    "tag:link",
			OrigError: errors.New("can't find config in public context"),
		}
	} else {
		config = configI.(*Config)
	}

	if l, ok := node.args["lang"]; ok {
		lv, err := l.Evaluate(linkContext)
		if err != nil {
			return err
		}
		lang = lv.String()
		if lang == "" {
			lang = "en"
		}
	}
	link := ""
	slug := ""
	id := ""
	switch node.what {
	case "top_categories":
		link = config.Routes.TopCategories
	case "top_content":
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
		searchQuery := ""
		if q, ok := node.args["query"]; ok {
			ql, err := q.Evaluate(linkContext)
			if err != nil {
				return err
			}
			if !ql.IsNil() {
				searchQuery = ql.String()
			}
			delete(copyArgs, "query")
		}
		link = strings.ReplaceAll(config.Routes.Search, ":query", url.PathEscape(searchQuery))
	case "model", "category", "channel", "content":
		if node.what == "model" {
			link = config.Routes.Model
		} else if node.what == "category" {
			link = config.Routes.Category
		} else if node.what == "channel" {
			link = config.Routes.Channel
		} else if node.what == "content" {
			link = config.Routes.ContentItem
		}
		if s, ok := node.args["slug"]; ok {
			se, err := s.Evaluate(linkContext)
			if err != nil {
				return err
			}
			if !se.IsNil() {
				slug = se.String()
				delete(copyArgs, "slug")
			}
		}
		if i, ok := node.args["id"]; ok {
			ie, err := i.Evaluate(linkContext)
			if err != nil {
				return err
			}
			if ie.IsString() {
				id = ie.String()
			} else if ie.IsInteger() {
				id = strconv.FormatInt(int64(ie.Integer()), 10)
			} else if !ie.IsNil() {
				id = fmt.Sprintf("%v", ie.Interface())
			}
			delete(copyArgs, "id")
		}
		if node.what == "content" {
			if s, ok := node.args["category"]; ok {
				se, err := s.Evaluate(linkContext)
				if err != nil {
					return err
				}
				category := se.String()
				delete(copyArgs, "category")
				link = strings.ReplaceAll(link, ":category", category)
			}
		}
		link = strings.ReplaceAll(link, ":slug", url.PathEscape(slug))
		link = strings.ReplaceAll(link, ":id", url.PathEscape(id))
	}
	if config.General.MultiLanguage {
		link = strings.ReplaceAll(config.Routes.LanguageTemplate, ":route", link)
		link = strings.ReplaceAll(link, ":lang", lang)
	}
	var pageNum int64 = 1
	if strings.Contains(link, ":page") {
		if s, ok := node.args["page"]; ok {
			se, err := s.Evaluate(linkContext)
			if err != nil {
				return err
			}
			page := fmt.Sprintf("%v", se.Interface())
			if page == "" || se.IsNil() {
				page = "1"
			}
			pageNum, _ = strconv.ParseInt(page, 10, 64)
			link = strings.ReplaceAll(link, ":page", page)
			delete(copyArgs, "page")
		}
	}
	isOut := false
	if s, ok := node.args["out"]; ok {
		out, err := s.Evaluate(linkContext)
		if err != nil {
			return err
		}
		if out.IsBool() && out.Bool() {
			var pageTemplate = ""
			if pageTemplate, ok = ctx.Public["page_template"].(string); !ok {
				log.Println("no page_template var found")
			}
			if node.what == "category" && pageTemplate == "top-categories" && pageNum == 1 {
				isOut = true
			}
			if node.what == "content" {
				if pageTemplate == "top-content" && pageNum == 1 {
					isOut = true
				} else if pageTemplate == "category" && pageNum == 1 {
					isOut = true
				}
			}
		}
		if out.IsBool() && out.Bool() && (node.what == "content" || node.what == "category") {
			isOut = true
		}
		delete(copyArgs, "out")
	}
	if isOut {
		// Link to out
		outLink := config.Routes.Out
		if config.General.MultiLanguage {
			outLink = strings.ReplaceAll(config.Routes.LanguageTemplate, ":route", outLink)
			outLink = strings.ReplaceAll(outLink, ":lang", lang)
		}
		outlinkParams := url.Values{}
		outlinkParams.Set(config.Params.CountRedirect, helpers.EncryptBase64(link))
		templateName := linkContext.Public["page_template"].(string)
		if node.what == "category" {
			outlinkParams.Set(config.Params.CountType, config.Params.CountTypeTopCategories)
			outlinkParams.Set(config.Params.CategoryId, id)
		} else if templateName == "top-content" {
			outlinkParams.Set(config.Params.CountType, config.Params.CountTypeTopContent)
			outlinkParams.Set(config.Params.ContentId, id)
		} else if templateName == "category" {
			category := linkContext.Public["category"].(*types.CategoryResult)
			outlinkParams.Set(config.Params.CategoryId, strconv.FormatInt(int64(category.Id), 10))
			outlinkParams.Set(config.Params.ContentId, id)
			outlinkParams.Set(config.Params.CountType, config.Params.CountTypeCategory)
		}
		outLink = outLink + "?" + outlinkParams.Encode()
		_, err := writer.WriteString(template.HTMLEscapeString(outLink))
		if err != nil {
			return &pongo2.Error{Sender: "tag:link", OrigError: err}
		}
		return nil
	}
	hasParams := false
	params := url.Values{}
	for key, v := range copyArgs {
		vv, err := v.Evaluate(linkContext)
		if err != nil {
			return err
		}
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
		case "duration_from":
			key = config.Params.DurationFrom
		case "duration_to":
			key = config.Params.DurationTo
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
		if s, ok := vv.Interface().(string); ok {
			// Parameter is string and key is sort_by, so we can replace some param values with user defined
			if key == config.Params.SortBy {
				switch s {
				case "sort_by_views":
					s = config.Params.SortByViews
				case "sort_by_duration":
					s = config.Params.SortByDuration
				case "sort_by_date":
					s = config.Params.SortByDate
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
			if s != "" {
				params.Add(key, s)
				hasParams = true
			}
		} else if !vv.IsNil() {
			val := fmt.Sprintf("%v", vv.Interface())
			if val != "" {
				params.Add(key, val)
				hasParams = true
			}
		}
	}
	if hasParams {
		link = link + "?" + params.Encode()
	}
	_, err := writer.WriteString(template.HTMLEscapeString(link))
	if err != nil {
		return &pongo2.Error{Sender: "tag:link", OrigError: err}
	}
	return nil
}

func pongo2Link(doc *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagLink := &tagLinkNode{
		args: make(map[string]pongo2.IEvaluator),
	}
	whatToken := arguments.MatchType(pongo2.TokenString)
	if whatToken == nil {
		return nil, arguments.Error("Expected string - one of: top_categories, top_content, autocomplete, search, popular, new, long, model, models, category, channel, content, out", nil)
	}
	whatTokenOk := false
	for _, l := range linksTokens {
		if l == whatToken.Val {
			whatTokenOk = true
			break
		}
	}
	if !whatTokenOk {
		return nil, arguments.Error("Expected string - one of: top_categories, top_content, autocomplete, search, popular, new, long, model, models, category, channel, content, out", nil)
	}
	tagLink.what = whatToken.Val
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
		tagLink.args[idToken.Val] = expression
	}
	return tagLink, nil
}
