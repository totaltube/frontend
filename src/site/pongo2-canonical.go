package site

import (
	"fmt"
	"net/http"
	"net/url"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v4"
)


func getCanonical(ctx pongo2.Context, page int64, q ...url.Values) string {
	if page == 0 {
		if p, ok := ctx["page"].(int64); ok {
			page = p
		}
	}
	config := ctx["config"].(*Config)
	var route string
	if r, ok := ctx["route"]; !ok {
		return ""
	} else {
		route = r.(string)
	}
	isSearchPage := route == config.Routes.Search
	alternateQuery := url.Values(http.Header(ctx["canonical_query"].(url.Values)).Clone())
	if strings.Contains(route, "{page}") {
		route = strings.ReplaceAll(route, "{page}", strconv.FormatInt(page, 10))
	} else if page > 1 {
		alternateQuery.Set(config.Params.Page, strconv.FormatInt(page, 10))
	}
	if params, ok := ctx["params"].(map[string]string); ok {
		for paramKey, paramValue := range params {
			route = strings.ReplaceAll(route, "{"+paramKey+"}", paramValue)
		}
	}
	if config.General.MultiLanguage && isSearchPage {
		// Для поисковой страницы нужно каноникал вместе с языком указывать.
		langId := ctx["lang"].(*types.Language).Id
		route = strings.ReplaceAll(config.Routes.LanguageTemplate, "{route}", route)
		route = strings.ReplaceAll(route, "{lang}", langId)
	}
	for _, qq := range q {
		for key, val := range qq {
			alternateQuery.Set(key, val[0])
		}
	}
	query := alternateQuery.Encode()
	if query != "" {
		route += "?" + query
	}
	return route
}

type tagCanonicalNode struct {
}

func (node *tagCanonicalNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	config := context.Public["config"].(*Config)
	hostName := context.Public["host"].(string)
	var page int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	route := getCanonical(context.Public, page)
	if config.General.CanonicalUrl != "" {
		route = strings.TrimSuffix(config.General.CanonicalUrl, "/")+route
	} else {
		route = "https://"+hostName+route
	}
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
