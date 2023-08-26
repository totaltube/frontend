package site

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"regexp"

	"github.com/dlclark/regexp2"
	"github.com/flosch/pongo2/v4"

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
	contextLang := lang
	var copyArgs = make(map[string]pongo2.IEvaluator)
	for k, v := range node.args {
		copyArgs[k] = v
	}
	var config *types.Config
	if configI, ok := linkContext.Public["config"]; !ok {
		return &pongo2.Error{
			Sender:    "tag:link",
			OrigError: errors.New("can't find config in public context"),
		}
	} else {
		config = configI.(*types.Config)
	}
	var changeLangLink bool
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
	if what == "current" {
		what = linkContext.Public["route"].(string)
		currentParams := linkContext.Public["params"].(map[string]string)
		for k, v := range currentParams {
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
	link = GetLink(what, config, lang, changeLangLink, args...)
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
