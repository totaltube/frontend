package site

import (
	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"
)

// repeat tag repeats any value in array given amount of times and saves it in given variable name.

type tagRepeatNode struct {
	what pongo2.IEvaluator
	amount pongo2.IEvaluator
	as string
}

func (node *tagRepeatNode) Execute(ctx *pongo2.ExecutionContext, _ pongo2.TemplateWriter) *pongo2.Error {
	// Evaluate what to repeat
	what, err := node.what.Evaluate(ctx)
	if err != nil {
		return err
	}
	var amount int
	if node.amount == nil {
		return &pongo2.Error{Sender: "tag:repeat", OrigError: errors.New("amount not set")}
	}
	if node.as == "" {
		return &pongo2.Error{Sender: "tag:repeat", OrigError: errors.New("as not set")}
	}
	a, err := node.amount.Evaluate(ctx)
	if err != nil {
		return err
	}
	amount = a.Integer()
	var result = make([]interface{}, 0, amount)
	for i := 0; i< amount; i ++ {
		result = append(result, what.Interface())
	}
	ctx.Public[node.as] = pongo2.AsValue(result)
	return nil
}

func pongo2Repeat(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &tagRepeatNode{}
	// parse what to repeat
	var err *pongo2.Error
	node.what, err = arguments.ParseExpression()
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
			idToken = arguments.MatchType(pongo2.TokenKeyword)
		}
		if idToken == nil {
			return nil, arguments.Error("Identifier or keyword expected", commaToken)
		}
		equalToken := arguments.MatchType(pongo2.TokenSymbol)
		if equalToken == nil || equalToken.Val != "=" {
			return nil, arguments.Error("= expected", idToken)
		}
		if idToken.Val == "as" {
			asToken := arguments.MatchType(pongo2.TokenIdentifier)
			if asToken == nil {
				return nil, arguments.Error("Identifier expected", asToken)
			}
			node.as = asToken.Val
		} else if idToken.Val == "amount" {
			expression, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.amount = expression
		} else {
			return nil, arguments.Error("unexpected param "+idToken.Val, idToken)
		}
	}
	return node, nil
}
