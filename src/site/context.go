package site

import (
	"github.com/flosch/pongo2/v4"
	"os"
	"path/filepath"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
	"time"
)

type alternateT struct {
	Lang string
	Url  string
}


func generateContext(name string, sitePath string, customContext pongo2.Context) pongo2.Context {
	var ctx = pongo2.Context{
		"translate": func(text string) string {
			return deferredTranslate("en", customContext["lang"].(*types.Language).Id, text)
		},
		"static": func(filePaths ...string) string {
			filePath := strings.Join(filePaths, "")
			p := filepath.Join(sitePath, "public", filePath)
			if fileInfo, err := os.Stat(p); err == nil {
				v := strconv.FormatInt(fileInfo.ModTime().Unix(), 10)
				v = v[len(v)-5:]
				return "/" + strings.TrimPrefix(filePath, "/") + "?v=" + v
			}
			return "/" + strings.TrimPrefix(filePath, "/")
		},
		"now": time.Now(),
	}
	return ctx.Update(customContext)
}
