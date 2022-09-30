package site

import (
	"html"
	"regexp"

	"github.com/flosch/pongo2/v4"
)

var replaceDynamicRegex = regexp.MustCompile(`<data class=["']?_dynamic["']? value=["']?([^"\\]*(?:\\.[^"\\]*)*)["']?></data>`)

func InsertDynamic(src []byte, userCtx pongo2.Context) (result []byte, err error) {
	result = replaceDynamicRegex.ReplaceAllFunc(src, func(bytes []byte) []byte {
		matches := replaceDynamicRegex.FindSubmatch(bytes)
		expression := html.UnescapeString(string(matches[1]))
		tpl, err := pongo2.FromString("{{" + expression + "}}")
		if err != nil {
			return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
		}
		result, err = tpl.ExecuteBytes(userCtx)
		if err != nil {
			return []byte("Error rendering dynamic expression [ " + expression + " ]: " + err.Error())
		}
		return result
	})
	userCtx["set_cookie"] = nil // unset dynamic function with fiber inside.
	return
}
