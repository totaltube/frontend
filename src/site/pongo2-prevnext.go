package site

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"sersh.com/totaltube/frontend/types"
)

type tagPrevnextNode struct {
}

func (node *tagPrevnextNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	langId := context.Public["lang"].(*types.Language).Id
	var page int64 = 1
	var pages int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	if p, ok := context.Public["pages"]; ok {
		pages = p.(int64)
	}
	if page > 1 {
		// there is a prev page
		_, err := writer.WriteString(fmt.Sprintf(`<link rel="prev" href="%s">`, getAlternate(context.Public, langId, page-1)))
		if err != nil {
			return &pongo2.Error{Sender: "tag:prevnext", OrigError: err}
		}
	}
	if page < pages {
		// there is a next page
		_, err := writer.WriteString(fmt.Sprintf(`<link rel="next" href="%s">`, getAlternate(context.Public, langId, page+1)))
		if err != nil {
			return &pongo2.Error{Sender: "tag:prevnext", OrigError: err}
		}
	}
	return nil
}

func pongo2Prevnext(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tag := &tagPrevnextNode{}
	return tag, nil
}
