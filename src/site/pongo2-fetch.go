package site

import (
	"errors"
	"github.com/flosch/pongo2/v4"
	"log"
	"sersh.com/totaltube/frontend/api"
)

type tagFetchNode struct {
	what    string
	wrapper *pongo2.NodeWrapper
	args    map[string]pongo2.IEvaluator
}

func (node *tagFetchNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	fetchContext := pongo2.NewChildExecutionContext(ctx)
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
	switch node.what {
	case "categories":
		results, err := api.CategoriesList(lang, int64(page), sort, int64(amount))
		if err != nil {
			log.Println(err)
		}
		fetchContext.Private["categories"] = results
	}
	err := node.wrapper.Execute(fetchContext, writer)
	return err
}

func pongo2Fetch(doc *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagFetch := &tagFetchNode{}
	var err *pongo2.Error
	whatToken := arguments.MatchType(pongo2.TokenString)
	if whatToken == nil {
		return nil, arguments.Error("Expected string - one of: categories, channels, searches, models ", nil)
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
		tagFetch.args[idToken.Val] = expression
	}
	tagFetch.wrapper, _, err = doc.WrapUntilTag("endfetch")
	if err != nil {
		return nil, err
	}
	return tagFetch, nil
}
