package site

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/flosch/pongo2/v4"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

var httpRegex = regexp.MustCompile(`(?i)^(https?://|//)`)
//language=Regexp
var paramRegex = regexp2.MustCompile(`\{([\w_]+)\}`, regexp2.None)

type tagLinkNode struct {
	what pongo2.IEvaluator
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
	var as string
	if asA, ok := node.args["as"]; ok {
		asAv, err := asA.Evaluate(linkContext)
		if err != nil {
			return err
		}
		as = asAv.String()
		delete(copyArgs, "as")
	}
	link := ""
	slug := ""
	id := ""
	w, err := node.what.Evaluate(linkContext)
	if err != nil {
		return err
	}
	isCustomRoute := false
	what := w.String()
	switch what {
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
		link = strings.ReplaceAll(config.Routes.Search, "{query}",
			strings.ReplaceAll(url.PathEscape(searchQuery), "%20", "+"))
	case "fake_player":
		link = config.Routes.FakePlayer
		if s, ok := node.args["slug"]; ok {
			ss, err := s.Evaluate(linkContext)
			if err != nil {
				return err
			}
			link = strings.ReplaceAll(link, "{slug}", ss.String())
			delete(copyArgs, "slug")
		}
		if s, ok := node.args["id"]; ok {
			ss, err := s.Evaluate(linkContext)
			if err != nil {
				return err
			}
			link = strings.ReplaceAll(link, "{id}", ss.String())
			delete(copyArgs, "id")
		}
	case "current":
		var ok bool
		if link, ok = linkContext.Public["route"].(string); !ok {
			log.Println("no route set")
		}
	case "model", "category", "channel", "content":
		if what == "model" {
			link = config.Routes.Model
		} else if what == "category" {
			link = config.Routes.Category
		} else if what == "channel" {
			link = config.Routes.Channel
		} else if what == "content" {
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
		if what == "content" {
			category := ""
			if s, ok := node.args["category"]; ok {
				se, err := s.Evaluate(linkContext)
				if err != nil {
					return err
				}
				category = se.String()
				delete(copyArgs, "category")
			}
			if s, ok := linkContext.Public["category"].(*types.CategoryResult); category == "" && ok {
				category = s.Slug
			}
			if s, ok := node.args["categories"]; ok {
				se, err := s.Evaluate(linkContext)
				if err != nil {
					return err
				}
				if ss, ok := se.Interface().(types.TaxonomyResults); ok {
					if len(ss) > 0 {
						category = []types.TaxonomyResult(ss)[0].Slug
					}
					delete(copyArgs, "categories")
				}
			}
			if category == "" {
				category = "porn"
			}
			link = strings.ReplaceAll(link, "{category}", category)
		}
		link = strings.ReplaceAll(link, "{slug}", url.PathEscape(slug))
		link = strings.ReplaceAll(link, "{id}", url.PathEscape(id))
	default:
		what = strings.TrimPrefix(what, "custom.")
		if r, ok := config.Routes.Custom[what]; ok {
			link = r
		} else {
			link = what
		}
		isCustomRoute = true
		if config.General.MultiLanguage  {
			link = strings.ReplaceAll(link, "{lang}", lang)
		}
		link, _ = paramRegex.ReplaceFunc(link, func(match regexp2.Match) string {
			if l, ok := node.args[match.Groups()[1].String()]; ok {
				delete(copyArgs, match.Groups()[1].String())
				lv, err := l.Evaluate(linkContext)
				if err != nil {
					log.Println(err)
					return match.String()
				}
				return url.PathEscape(lv.String())
			}
			return match.String()
		}, -1, -1)
	}
	if config.General.MultiLanguage && !httpRegex.MatchString(link) && !isCustomRoute {
		link = strings.ReplaceAll(config.Routes.LanguageTemplate, "{route}", link)
		link = strings.ReplaceAll(link, "{lang}", lang)
	}
	var pageNum int64 = 1
	if strings.Contains(link, "{page}") {
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
			link = strings.ReplaceAll(link, "{page}", page)
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
			if what == "category" && pageTemplate == "top-categories" && pageNum == 1 {
				isOut = true
			}
			if what == "content" {
				if pageTemplate == "top-content" && pageNum == 1 {
					isOut = true
				} else if pageTemplate == "category" && pageNum == 1 {
					isOut = true
				}
			}
		}
		if out.IsBool() && out.Bool() && (what == "content" || what == "category") {
			isOut = true
		}
		delete(copyArgs, "out")
	}
	isTrade := false
	if s, ok := node.args["with_trade"]; ok {
		trade, err := s.Evaluate(linkContext)
		if err != nil {
			return err
		}
		isTrade = trade.Bool()
		delete(copyArgs, "with_trade")
	}
	if isOut {
		// Link to out
		outLink := config.Routes.Out
		outlinkParams := url.Values{}
		outlinkParams.Set(config.Params.CountRedirect, helpers.EncryptBase64(link))
		templateName := linkContext.Public["page_template"].(string)
		if what == "category" {
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
		if isTrade {
			outLink = strings.ReplaceAll(outLink, "{{encoded_url}}",
				strings.ReplaceAll(config.General.TradeUrlTemplate, "{{url}}", outLink))
		}
		_, err := writer.WriteString(template.HTMLEscapeString(outLink))
		if err != nil {
			return &pongo2.Error{Sender: "tag:link", OrigError: err}
		}
		return nil
	}
	params := url.Values{}
	if what == "current" {
		// Setting current uri params
		if uriParams, ok := linkContext.Public["params"].(map[string]string); ok {
			for paramKey, paramValue := range uriParams {
				link = strings.ReplaceAll(link, "{"+paramKey+"}", paramValue)
			}
		}
		// Copying current query params
		params = url.Values(http.Header(linkContext.Public["canonical_query"].(url.Values)).Clone())
	}
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
		if s, ok := vv.Interface().(string); ok {
			// Parameter is string and key is sort_by, so we can replace some param values with user defined
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
				params.Del(key)
			} else {
				params.Set(key, s)
			}
		} else if !vv.IsNil() {
			val := fmt.Sprintf("%v", vv.Interface())
			if val != "" {
				params.Set(key, val)
			}
		}
	}
	if len(params) > 0 {
		link = link + "?" + params.Encode()
	}
	if isTrade {
		link = strings.ReplaceAll(link, "{{encoded_url}}",
			strings.ReplaceAll(config.General.TradeUrlTemplate, "{{url}}", link))
	}
	if as != "" {
		linkContext.Public[as] = link
	} else {
		_, err1 := writer.WriteString(template.HTMLEscapeString(link))
		if err1 != nil {
			return &pongo2.Error{Sender: "tag:link", OrigError: err1}
		}
	}
	return nil
}

func pongo2Link(doc *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagLink := &tagLinkNode{
		args: make(map[string]pongo2.IEvaluator),
	}
	var err *pongo2.Error
	tagLink.what, err = arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
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
