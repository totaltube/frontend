package site

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/dlclark/regexp2"
	"github.com/flosch/pongo2/v6"

	"sersh.com/totaltube/frontend/types"
)

var httpRegex = regexp.MustCompile(`(?i)^(https?://|//)`)

// language=Regexp
var paramRegex = regexp2.MustCompile(`\{([\w_]+)\}`, regexp2.None)

type tagLinkNode struct {
	what pongo2.IEvaluator
	args map[string]pongo2.IEvaluator
}

func (node *tagLinkNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	linkContext := pongo2.NewChildExecutionContext(ctx)
	hostName := linkContext.Public["host"].(string)
	var config *types.Config
	if configI, ok := linkContext.Public["config"]; !ok {
		return &pongo2.Error{
			Sender:    "tag:link",
			OrigError: errors.New("can't find config in public context"),
		}
	} else {
		config = configI.(*types.Config)
	}
	// do not override hostName with default here; keep real request host
	lang := "en"

	if l, ok := linkContext.Public["lang"]; ok {
		lang = l.(*types.Language).Id
	}
	contextLang := lang
	var copyArgs = make(map[string]pongo2.IEvaluator)
	for k, v := range node.args {
		copyArgs[k] = v
	}

	var changeLangLink bool // if true, then we need to change lang in link
	if l, ok := node.args["lang"]; ok {
		lv, err := l.Evaluate(linkContext)
		if err != nil {
			return err
		}
		lang = lv.String()
		if lang == "" {
			lang = "en"
		}
		delete(copyArgs, "lang")
		if lang != contextLang {
			changeLangLink = true
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
	w, err := node.what.Evaluate(linkContext)
	if err != nil {
		return err
	}
	what := w.String()
	var args = make([]interface{}, 0, 30)
	for name, a := range copyArgs {
		value, err := a.Evaluate(linkContext)
		if err != nil {
			return err
		}
		args = append(args, name, value.Interface())
	}
	if d, ok := config.LanguageDomains[lang]; ok && d != "" {
		// target language has its own domain -> force absolute to that domain
		hostName = d
		args = append(args, "full_url", true)
	} else if def, ok := config.LanguageDomains["default"]; ok && def != "" && hostName != def {
		// target language has no dedicated domain, but current host is not default ->
		// force absolute to default domain
		hostName = def
		args = append(args, "full_url", true)
	}
	if what == "current" {
		what = linkContext.Public["page_template"].(string)
		if what == "search" && contextLang != lang {
			// For search pages we can't change the lang in the link, because it will change the search results, so we will redirect to another page
			var found = false
			for k, v := range config.Routes.Custom {
				if v == "/" {
					what = "custom." + k
					found = true
					break
				}
			}
			if !found && config.Routes.TopCategories == "/" {
				what = "top_categories"
				found = true
			}
			if !found && config.Routes.New == "/" {
				what = "new"
				found = true
			}
			if !found && config.Routes.Popular == "/" {
				what = "popular"
				found = true
			}
			if !found && config.Routes.TopContent == "/" {
				what = "top_content"
				found = true
			}
			if !found {
				if config.Routes.TopCategories == "/" {
					what = "top_categories"
				} else if config.Routes.New != "" && config.Routes.New != "-" {
					what = "new"
				} else if config.Routes.Popular != "" && config.Routes.Popular != "-" {
					what = "popular"
				} else {
					what = "top_content"
				}
			}
		} else {
			currentParams := linkContext.Public["params"].(map[string]string)
			for k, v := range currentParams {
				if k == "id" && config.Routes.IdXorKey > 0 {
					numericId, _ := strconv.ParseInt(v, 10, 64)
					if numericId > 0 {
						numericId = numericId ^ config.Routes.IdXorKey
					}
					v = strconv.FormatInt(numericId, 10)
				}
				args = append(args, k, v)
			}
			queryParams := url.Values(http.Header(linkContext.Public["canonical_query"].(url.Values)).Clone())
			for k, v := range queryParams {
				for _, vv := range v {
					// prepending query params, because they can be overwritten by template params
					args = append([]interface{}{k, vv}, args...)
				}
			}
		}
	}
	link = GetLink(what, config, hostName, lang, changeLangLink, args...)
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
