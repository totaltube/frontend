package site

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"
	"sersh.com/totaltube/frontend/types"
)

type tagPageLinkNode struct {
	page pongo2.IEvaluator
}

func (node *tagPageLinkNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	langId := "en"
	if l, ok := context.Public["lang"]; ok {
		langId = l.(*types.Language).Id
	}
	v, err := node.page.Evaluate(context)
	if err != nil {
		return err
	}
	if !v.IsInteger() {
		return &pongo2.Error{
			Sender: "tag:page_link",
			OrigError: errors.New(fmt.Sprintf("page must be integer, %T given", v.Interface())),
		}
	}
	page := v.Integer()
	if page < 1 {
		page = 1
	}
	_, err1 := writer.WriteString(getAlternate(context.Public, langId, int64(page)))
	if err1 != nil {
		return &pongo2.Error{Sender: "tag:page_link", OrigError: err1}
	}
	return nil
}

func pongo2PageLink(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tag := &tagPageLinkNode{}
	expression, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	tag.page = expression
	return tag, nil
}

