package site

import (
	"github.com/flosch/pongo2/v4"
	"os"
	"path/filepath"
	"sersh.com/totaltube/frontend/helpers"
	"strconv"
	"strings"
)

func generateContext(name string, sitePath string, config *Config, customContext pongo2.Context) pongo2.Context {
	var ctx = pongo2.Context{
		"translate": func(text string) string {
			return deferredTranslate("en", customContext["lang"].(string), text)
		},
		"static": func(filePath string) string {
			p := filepath.Join(sitePath, "public", filePath)
			if fileInfo, err := os.Stat(p); err == nil {
				v := helpers.Md5Hash(strconv.FormatInt(int64(fileInfo.ModTime().Nanosecond()), 10))[0:4]
				return "/" + strings.TrimPrefix(filePath, "/") + "?v="+v
			}
			return "/" + strings.TrimPrefix(filePath, "/")
		},
	}
	return ctx.Update(customContext)
}
