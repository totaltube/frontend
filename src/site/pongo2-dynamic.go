package site

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v6"
	"sersh.com/totaltube/frontend/helpers"
)

type tagNocacheNode struct {
	expression       string
	inner_expression string
}

func (node *tagNocacheNode) Execute(_ *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	var err error
	if node.inner_expression != "" {
		_, err = writer.WriteString(fmt.Sprintf(`<data class="_dynamic" value="inner">%s</data>`, helpers.Base64([]byte(node.inner_expression))))
		if err != nil {
			return &pongo2.Error{Sender: "tag:dynamic", OrigError: err}
		}
	} else {

		_, err = writer.WriteString(fmt.Sprintf(`<data class="_dynamic" value="short">%s</data>`, helpers.Base64([]byte(node.expression))))
		if err != nil {
			return &pongo2.Error{Sender: "tag:dynamic", OrigError: err}
		}
	}
	return nil
}

func pongo2Dynamic(doc *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
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
	if strings.TrimSpace(tagNocache.expression) == "" {
		// Process tokens until the "enddynamic" tag
		var err *pongo2.Error
		tagNocache.inner_expression, err = extractTextUntilTag(doc, "enddynamic")
		if err != nil {
			return nil, err
		}
	}
	return tagNocache, nil
}

func extractTextUntilTag(p *pongo2.Parser, names ...string) (string, *pongo2.Error) {
	var resultText string

	for p.Remaining() > 0 {
		// Check if there is an opening symbol '{%'
		if p.Peek(pongo2.TokenSymbol, "{%") != nil {
			// Get the tag identifier that comes after '{%'
			tagIdent := p.PeekTypeN(1, pongo2.TokenIdentifier)

			if tagIdent != nil {
				// Check if the identifier matches one of the desired tags
				found := false
				for _, n := range names {
					if tagIdent.Val == n {
						found = true
						break
					}
				}

				// If the end of the desired tag is found, stop collecting text
				if found {
					// Skip the '{%' tokens and the tag name
					p.ConsumeN(2)
					// Skip tokens until the closing '%}'
					for {
						if p.Match(pongo2.TokenSymbol, "%}") != nil {
							return resultText, nil
						}
						token := p.Current()
						p.Consume()
						if token == nil {
							return "", p.Error("Unexpected EOF.", p.Current())
						}
					}
				}
			}
		}

		// Add the text of the current token to the result
		token := p.Current()
		if token != nil {
			resultText += token.Val
			p.Consume()
		} else {
			return "", p.Error("Unexpected EOF.", token)
		}
	}

	return "", p.Error(fmt.Sprintf("Unexpected EOF, expected tag %s.", strings.Join(names, " or ")), p.Current())
}
