package site

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"sersh.com/totaltube/frontend/types"
)

type tagAlternatesNode struct {
}

func (node *tagAlternatesNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	config := context.Public["config"].(*Config)
	if !config.General.MultiLanguage {
		return nil
	}
	langId := context.Public["lang"].(*types.Language).Id
	languages := context.Public["languages"].([]types.Language)
	var page int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	for _, l := range languages {
		if l.Id == langId {
			continue
		}
		_, err := writer.WriteString(
			fmt.Sprintf(`<link rel="alternate" hreflang="%s" href="%s">`+"\n",
				l.Id, getCanonical(context.Public, l.Id, page)))
		if err != nil {
			return &pongo2.Error{Sender: "tag:alternates", OrigError: err}
		}
	}
	return nil
}

func pongo2Alternates(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tag := &tagAlternatesNode{}
	return tag, nil
}

