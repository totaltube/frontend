package site

import (
	"errors"
	"github.com/flosch/pongo2/v6"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"reflect"
	"sersh.com/totaltube/frontend/types"
	"sort"
)

type sortNode struct {
	what pongo2.IEvaluator
	args map[string]pongo2.IEvaluator
}

func (node *sortNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	sortContext := pongo2.NewChildExecutionContext(ctx)
	langTag := language.English
	if l, ok := sortContext.Public["lang"]; ok {
		langTag = l.(*types.Language).Tag
	}
	items, err := node.what.Evaluate(sortContext)
	if err != nil {
		return err
	}
	if !items.CanSlice() {
		return &pongo2.Error{
			Sender:    "tag:sort",
			OrigError: errors.New("wrong items param, not array"),
		}
	}
	var sortType = "title"
	if t, ok := node.args["type"]; ok {
		tp, err := t.Evaluate(sortContext)
		if err != nil {
			return err
		}
		sortType = tp.String()
	}
	cc := collate.New(langTag, collate.Loose)
	switch it := items.Interface().(type) {
	case []*types.CategoryResult:
		if sortType == "title" {
			sort.Slice(it, func(i, j int) bool {
				return cc.CompareString(it[i].Title, it[j].Title) == -1
			})
		} else if sortType == "total" || sortType == "total_desc" {
			sort.Slice(it, func(i, j int) bool {
				return it[i].Total > it[j].Total
			})
		} else if sortType == "total_asc" {
			sort.Slice(it, func(i, j int) bool {
				return it[i].Total < it[j].Total
			})
		}
	case []*types.ModelResult:
		if sortType == "title" {
			sort.Slice(it, func(i, j int) bool {
				return cc.CompareString(it[i].Title, it[j].Title) == -1
			})
		} else if sortType == "total" || sortType == "total_desc" {
			sort.Slice(it, func(i, j int) bool {
				return it[i].Total > it[j].Total
			})
		} else if sortType == "total_asc" {
			sort.Slice(it, func(i, j int) bool {
				return it[i].Total < it[j].Total
			})
		}
	case []*types.ChannelResult:
		if sortType == "title" {
			sort.Slice(it, func(i, j int) bool {
				return cc.CompareString(it[i].Title, it[j].Title) == -1
			})
		} else if sortType == "total" || sortType == "total_desc" {
			sort.Slice(it, func(i, j int) bool {
				return it[i].Total > it[j].Total
			})
		} else if sortType == "total_asc" {
			sort.Slice(it, func(i, j int) bool {
				return it[i].Total < it[j].Total
			})
		}
	default:
		return &pongo2.Error{
			Sender:    "tag:sort",
			OrigError: errors.New("unknown array type to sort: " + reflect.TypeOf(items.Interface()).String()),
		}
	}
	return nil
}

func pongo2Sort(doc *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	sortN := &sortNode{
		args: make(map[string]pongo2.IEvaluator),
	}
	what, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	sortN.what = what
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
		sortN.args[idToken.Val] = expression
	}
	return sortN, nil
}
