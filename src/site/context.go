package site

import (
	"github.com/flosch/pongo2/v4"
)

func generateContext(name string, config *Config, customContext pongo2.Context) pongo2.Context {
	var ctx = pongo2.Context{
		"translate": func(text string) string {
			return deferredTranslate("en", customContext["lang"].(string), text)
		},
	}
	return ctx.Update(customContext)
}
