package site

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/types"

	"github.com/flosch/pongo2/v6"
)

func getCanonical(ctx pongo2.Context, page int64, q ...url.Values) string {
	if page == 0 {
		if p, ok := ctx["page"].(int64); ok {
			page = p
		}
	}
	config := ctx["config"].(*types.Config)
	langId := ctx["lang"].(*types.Language).Id
	hostName := ctx["host"].(string)
	var route string
	if r, ok := ctx["route"]; !ok {
		return ""
	} else {
		route = r.(string)
	}
	if route == config.Routes.VideoEmbed || route == config.Routes.FakePlayer {
		route = config.Routes.ContentItem
	}
	//isSearchPage := route == config.Routes.Search
	alternateQuery := url.Values(http.Header(ctx["canonical_query"].(url.Values)).Clone())
	if page > 1 {
		route = paginationRoute(route, config)
	}
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
	var canonicalPrefix string
	if config.General.CanonicalUrl != "" {
		canonicalPrefix = strings.TrimSuffix(config.General.CanonicalUrl, "/")
	} else {
		canonicalPrefix = "https://" + hostName
	}
	replacedLang := false
	if config.General.MultiLanguage && config.LanguageDomains[langId] != "" {
		canonicalPrefix = "https://" + config.LanguageDomains[langId]
		route = strings.ReplaceAll(config.Routes.LanguageTemplate, "{route}", route)
		route = strings.ReplaceAll(route, "{lang}", langId)
		replacedLang = true
	}
	if config.General.MultiLanguage && (langId != config.General.DefaultLanguage || !config.General.NoRedirectDefaultLanguage) && !replacedLang {
		// add language to the route if it's not the default language
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
	return canonicalPrefix + route
}

type tagCanonicalNode struct {
}

func (node *tagCanonicalNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	var page int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	route := getCanonical(context.Public, page)
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
