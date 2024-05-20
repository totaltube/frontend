package site

import "github.com/flosch/pongo2/v6"

// a little modified set tag. Make use of public context, not the private.

type tagSetNode struct {
	name       string
	expression pongo2.IEvaluator
}

func (node *tagSetNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	// Evaluate expression
	value, err := node.expression.Evaluate(ctx)
	if err != nil {
		return err
	}
	ctx.Public[node.name] = value
	return nil
}

func pongo2Set(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &tagSetNode{}

	// Parse variable name
	typeToken := arguments.MatchType(pongo2.TokenIdentifier)
	if typeToken == nil {
		return nil, arguments.Error("Expected an identifier.", nil)
	}
	node.name = typeToken.Val

	if arguments.Match(pongo2.TokenSymbol, "=") == nil {
		return nil, arguments.Error("Expected '='.", nil)
	}
	// Variable expression
	keyExpression, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	node.expression = keyExpression

	// Remaining arguments
	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed 'set'-tag arguments.", nil)
	}

	return node, nil
}
