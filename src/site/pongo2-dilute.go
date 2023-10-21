package site

import (
	"log"
	"math"
	"math/rand"

	"github.com/flosch/pongo2/v4"
	"github.com/pkg/errors"
)


type tagDiluteNode struct {
	what pongo2.IEvaluator // what array to dilute
	with pongo2.IEvaluator // dilute with this array
	from pongo2.IEvaluator // dilute from this index
	to   pongo2.IEvaluator // dilute to this index
	max  pongo2.IEvaluator // max items to take for dilute
	as   string            // save new array as this variable name
}

func (node *tagDiluteNode) Execute(ctx *pongo2.ExecutionContext, tw pongo2.TemplateWriter) *pongo2.Error {
	// Evaluate what to repeat
	what, err := node.what.Evaluate(ctx)
	if err != nil {
		return err
	}
	if !what.CanSlice() {
		return &pongo2.Error{Sender: "tag:dilute" ,OrigError: errors.New("source for dilute must be array")}
	}
	sourceLength := what.Len()
	with, err := node.with.Evaluate(ctx)
	if err != nil {
		return err
	}
	if !with.CanSlice() {
		log.Printf("%T, %v", with.Interface(), with.Interface())
		return &pongo2.Error{Sender: "tag:dilute", OrigError: errors.New("variable to dilute with must be array")}
	}
	diluteLength := with.Len()
	from := 0
	to := sourceLength
	max := diluteLength
	if node.from != nil {
		f, err := node.from.Evaluate(ctx)
		if err != nil {
			return err
		}
		from = f.Integer()
	}
	if node.to != nil {
		t, err := node.to.Evaluate(ctx)
		if err != nil {
			return err
		}
		to = t.Integer()
	}
	if node.max != nil {
		m, err := node.max.Evaluate(ctx)
		if err != nil {
			return err
		}
		max = m.Integer()
	}
	if node.as == "" {
		return &pongo2.Error{Sender: "tag:dilute", OrigError: errors.New("as not set")}
	}
	var result = make([]interface{}, 0, sourceLength+max)
	var itemsToDilute = make([]interface{}, 0, int(math.Max(float64(sourceLength+max-from), 0)))
	var itemsLeft = make([]interface{}, 0, int(math.Max(float64(sourceLength-to), 0)))
	what.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
		if idx < from {
			result = append(result, key.Interface())
		} else if idx < to {
			itemsToDilute = append(itemsToDilute, key.Interface())
		} else {
			itemsLeft = append(itemsLeft, key.Interface())
		}
		return true
	}, func() {})
	// inserting our
	with.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
		if idx >= max {
			return true
		}
		indexToInsert := int(math.Floor(rand.Float64() * float64(len(itemsToDilute))))
		if indexToInsert+1 > len(itemsToDilute) {
			return true
		}
		itemsToDilute = append(itemsToDilute[:indexToInsert+1], itemsToDilute[indexToInsert:]...)
		itemsToDilute[indexToInsert] = key.Interface()
		return true
	}, func() {})
	result = append(result, itemsToDilute...)
	result = append(result, itemsLeft...)
	ctx.Public[node.as] = pongo2.AsValue(result)
	return nil
}

func pongo2Dilute(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &tagDiluteNode{}
	// parse what to dilute
	var err *pongo2.Error
	node.what, err = arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	commaToken := arguments.Match(pongo2.TokenSymbol, ",")
	if commaToken == nil {
		return nil, arguments.Error("expecting comma", commaToken)
	}
	node.with, err = arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	for {
		if arguments.Match(pongo2.TokenSymbol, ",") == nil {
			break
		}
		idToken := arguments.MatchType(pongo2.TokenIdentifier)
		if idToken == nil {
			idToken = arguments.MatchType(pongo2.TokenKeyword)
		}
		if idToken == nil {
			return nil, arguments.Error("Identifier or keyword expected", commaToken)
		}
		equalToken := arguments.Match(pongo2.TokenSymbol, "=")
		if equalToken == nil {
			return nil, arguments.Error("= expected", equalToken)
		}
		switch idToken.Val {
		case "as":
			asToken := arguments.MatchType(pongo2.TokenIdentifier)
			if asToken == nil {
				return nil, arguments.Error("identifier expected", asToken)
			}
			node.as = asToken.Val
		case "from":
			fromToken, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.from = fromToken
		case "to":
			toToken, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.to = toToken
		case "max":
			maxToken, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.max = maxToken
		default:
			return nil, arguments.Error("unexpected identifier "+idToken.Val, idToken)
		}
	}
	return node, nil
}
