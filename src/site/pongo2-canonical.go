package site

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"net/http"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
)


func getCanonical(ctx pongo2.Context, langId string, page int64) string {
	config := ctx["config"].(*Config)
	var route string
	if r, ok := ctx["route"]; !ok {
		return ""
	} else {
		route = r.(string)
	}
	canonicalQuery := url.Values(http.Header(ctx["canonical_query"].(url.Values)).Clone())
	if strings.Contains(route, ":page") {
		route = strings.ReplaceAll(route, ":page", strconv.FormatInt(page, 10))
	} else if page > 1 {
		canonicalQuery.Set(config.Params.Page, strconv.FormatInt(page, 10))
	}
	if config.General.MultiLanguage {
		route = strings.ReplaceAll(config.Routes.LanguageTemplate, ":route", route)
		route = strings.ReplaceAll(route, ":lang", langId)
	}
	query := canonicalQuery.Encode()
	if query != "" {
		route += "?" + query
	}
	return route
}

type tagCanonicalNode struct {
}

func (node *tagCanonicalNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	langId := context.Public["lang"].(*types.Language).Id
	var page int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	route := getCanonical(context.Public, langId, page)
	_, err := writer.WriteString(fmt.Sprintf(`<link rel="canonical" href="%s">`, route))
	if err != nil {
		return &pongo2.Error{Sender: "tag:canonical", OrigError: err}
	}
	return nil
}

func pongo2Canonical(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagCanonical := &tagCanonicalNode{}
	return tagCanonical, nil
}
