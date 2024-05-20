package site

import (
	"bytes"
	"github.com/beevik/etree"
	"github.com/flosch/pongo2/v6"
)

// a little modified set tag. Make use of public context, not the private.

type tagCDataNode struct {
	expression pongo2.IEvaluator
}

func (node *tagCDataNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	// Evaluate expression
	value, err := node.expression.Evaluate(ctx)
	if err != nil {
		return err
	}
	cdata := etree.NewCData(value.String())
	buf := bytes.NewBuffer(nil)
	cdata.WriteTo(buf, &etree.WriteSettings{})
	_, err1 := writer.Write(buf.Bytes())
	if err1 != nil {
		return &pongo2.Error{Sender: "tag:cdata", OrigError: err}
	}
	return nil
}

func pongo2CData(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &tagCDataNode{}
	expression, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	node.expression = expression
	return node, nil
}
