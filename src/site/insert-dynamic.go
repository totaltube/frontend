package site

import (
	"errors"
	"html"
	"log"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v6"
	"sersh.com/totaltube/frontend/helpers"
)

var replaceDynamicRegex = regexp.MustCompile(`<data class=["']?_dynamic["']? value=["']?(.*?)["']?/?>(.*?)</data>`)

func InsertDynamic(src []byte, path string, userCtx pongo2.Context) (result []byte) {
	result = replaceDynamicRegex.ReplaceAllFunc(src, func(match []byte) (result []byte) {
		var err error
		matches := replaceDynamicRegex.FindSubmatch(match)
		expression := html.UnescapeString(string(matches[1]))
		if strings.HasPrefix(expression, "include ") {
			var tpl *pongo2.Template
			tpl, err = pongo2.FromString("{{" + strings.TrimPrefix(expression, "include ") + "}}")
			if err != nil {
				return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
			}
			var templateName string
			templateName, err = tpl.Execute(userCtx)
			if err != nil {
				return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
			}
			sp := strings.Split(templateName, ".")
			if len(sp) > 1 && sp[len(sp)-1] == "twig" {
				sp = sp[0 : len(sp)-1]
			}
			tpl, err = GetTemplate(strings.Join(sp, "."), path)
			if err != nil {
				if err == ErrTemplateNotFound {
					err = errors.New("wrong template name")
				}
				return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
			}
			result, err = tpl.ExecuteBytes(userCtx)
			if err != nil {
				log.Println(err)
			}
			return result
		}
		if expression != "" && expression != "inner" {
			var tpl *pongo2.Template
			tpl, err = pongo2.FromString("{{" + expression + "}}")
			if err != nil {
				return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
			}
			result, err = tpl.ExecuteBytes(userCtx)
			if err != nil {
				return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
			}
		}
		innerExpression := matches[2]
		if string(innerExpression) != "" {
			templateCode, err := helpers.FromBase64(string(innerExpression))
			if err != nil {
				log.Println("Error rendering dynamic expression [ " + string(innerExpression) + " ]: " + err.Error())
				return []byte("")
			}
			var tpl *pongo2.Template
			tpl, err = pongo2.FromBytes(templateCode)
			if err != nil {
				return []byte("Error rendering dynamic expression [ " + string(templateCode) + " ]: " + err.Error())
			}
			result, err = tpl.ExecuteBytes(userCtx)
			if err != nil {
				return []byte("Error rendering dynamic expression [ " + string(templateCode) + " ]: " + err.Error())
			}
		}
		return result
	})
	return
}
