package site

import (
	"github.com/flosch/pongo2/v4"
	"sersh.com/totaltube/frontend/types"
)

type tagAlternateNode struct {
	lang pongo2.IEvaluator
}

func (node *tagAlternateNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	config := context.Public["config"].(*Config)
	langId := context.Public["lang"].(*types.Language).Id
	var page int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	if !config.General.MultiLanguage {
		_, err := writer.WriteString(getCanonical(context.Public, langId, page))
		if err != nil {
			return &pongo2.Error{Sender: "tag:alternate", OrigError: err}
		}
		return nil
	}
	v, err := node.lang.Evaluate(context)
	if err != nil {
		return err
	}
	alternateLang := v.String()
	_, err1 := writer.WriteString(getCanonical(context.Public, alternateLang, page))
	if err1 != nil {
		return &pongo2.Error{Sender: "tag:alternate", OrigError: err1}
	}
	return nil
}

func pongo2Alternate(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tag := &tagAlternateNode{}
	expression, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	tag.lang = expression
	return tag, nil
}

