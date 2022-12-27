package site

import (
	"fmt"
	"html"

	"github.com/flosch/pongo2/v4"
)

type tagNocacheNode struct {
	expression string
}

func (node *tagNocacheNode) Execute(_ *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	var err error
	_, err = writer.WriteString(fmt.Sprintf(`<data class="_dynamic" value="%s"></data>`, html.EscapeString(node.expression)))
	if err != nil {
		return &pongo2.Error{Sender: "tag:dynamic", OrigError: err}
	}
	return nil
}

func pongo2Dynamic(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagNocache := &tagNocacheNode{}
	var amountTokens = arguments.Remaining()
	for k := 0; k < amountTokens; k++ {
		token := arguments.Get(k)
		add := token.Val
		if add == "include" {
			add = "include "
		}
		if token.Typ == pongo2.TokenString {
			add = `"` + add + `"`
		}
		tagNocache.expression += add
	}
	return tagNocache, nil
}
