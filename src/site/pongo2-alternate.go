package site

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"

	"sersh.com/totaltube/frontend/types"
)

type tagAlternateNode struct {
	lang pongo2.IEvaluator
}

func getAlternate(ctx pongo2.Context, langId string, page int64, q ...url.Values) string {
	if langId == "" {
		langId = ctx["lang"].(*types.Language).Id
	}
	if page == 0 {
		if p, ok := ctx["page"].(int64); ok {
			page = p
		}
	}
	config := ctx["config"].(*types.Config)
	var route string
	if r, ok := ctx["route"]; !ok {
		return ""
	} else {
		route = r.(string)
	}
	langTemplate := config.Routes.LanguageTemplate
	if ctn, ok := ctx["custom_template_name"].(string); ok {
		if customLangTemplate, exists := config.Routes.Custom[ctn+"_multilang"]; exists {
			langTemplate = customLangTemplate
		}
	}

	if route == config.Routes.VideoEmbed || route == config.Routes.FakePlayer {
		route = config.Routes.ContentItem
	}
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
	if config.General.MultiLanguage && (config.General.DefaultLanguage != langId || !config.General.NoRedirectDefaultLanguage) {
		route = strings.ReplaceAll(langTemplate, "{route}", route)
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
func (node *tagAlternateNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	config := context.Public["config"].(*types.Config)
	hostName := context.Public["host"].(string)
	langId := context.Public["lang"].(*types.Language).Id
	var page int64 = 1
	if p, ok := context.Public["page"]; ok {
		page = p.(int64)
	}
	var canonicalUrl = strings.TrimSuffix(config.General.CanonicalUrl, "/")
	if canonicalUrl == "" {
		canonicalUrl = "https://"+hostName
	}
	if !config.General.MultiLanguage {
		_, err := writer.WriteString(canonicalUrl+ getAlternate(context.Public, langId, page))
		if err != nil {
			return &pongo2.Error{Sender: "tag:alternate", OrigError: err}
		}
		return nil
	}
	v, err := node.lang.Evaluate(context)
	if err != nil {
		return &pongo2.Error{Sender: "tag:alternate", OrigError: err}
	}
	alternateLang := v.String()
	if templateName, ok := context.Public["page_template"].(string); ok && templateName == "search" {
		// Search page doesn't have alternate. Return link to the root
		link := strings.ReplaceAll(config.Routes.LanguageTemplate, "{lang}", alternateLang)
		link = strings.ReplaceAll(link, "{route}", "/")

		_, err := writer.WriteString(canonicalUrl+link)
		if err != nil {
			return &pongo2.Error{Sender: "tag:alternate", OrigError: err}
		}
		return nil
	}
	_, err1 := writer.WriteString(canonicalUrl+ getAlternate(context.Public, alternateLang, page))
	if err1 != nil {
		return &pongo2.Error{Sender: "tag:alternate", OrigError: err1}
	}
	return nil
}

func pongo2Alternate(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tag := &tagAlternateNode{}
	expression, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	tag.lang = expression
	return tag, nil
}

