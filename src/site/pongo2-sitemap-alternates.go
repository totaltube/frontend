package site

import (
	"fmt"

	"github.com/flosch/pongo2/v6"

	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

type tagSitemapAlternatesNode struct{}

func (node *tagSitemapAlternatesNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	context := pongo2.NewChildExecutionContext(ctx)
	config := context.Public["config"].(*types.Config)
	if !config.General.MultiLanguage {
		return nil
	}

	hostName := context.Public["host"].(string)
	if d, ok := internal.GetDefaultLanguageDomainValue(config); ok && d != "" {
		hostName = d
	}

	// content_item required in sitemap-video template
	contentItem, ok := context.Public["content_item"].(*types.ContentResult)
	if !ok || contentItem == nil {
		// Nothing to render if no content
		return nil
	}

	// Build default language link (self + x-default)
	defaultLang := config.General.DefaultLanguage
	defaultHost := hostName
	if d, ok := config.LanguageDomains[defaultLang]; ok && d != "" {
		defaultHost = d
	}
	defaultHref := GetLink(
		"content_item",
		config,
		defaultHost,
		defaultLang,
		false,
		"slug", contentItem.Slug,
		"id", contentItem.Id,
		"categories", contentItem.Categories,
		"full_url", true,
	)
	if defaultHref != "" {
		if _, err := fmt.Fprintf(writer, `<xhtml:link rel="alternate" hreflang="%s" href="%s"/></xhtml:link>`, defaultLang, defaultHref); err != nil {
			return &pongo2.Error{Sender: "tag:sitemap_alternates", OrigError: err}
		}
		if _, err := fmt.Fprintf(writer, `<xhtml:link rel="alternate" hreflang="x-default" href="%s"/></xhtml:link>`, defaultHref); err != nil {
			return &pongo2.Error{Sender: "tag:sitemap_alternates", OrigError: err}
		}
	}

	// Render alternates for other languages available in sitemap
	for _, lang := range internal.GetLanguagesAvailableInSitemap(config) {
		if lang.Id == defaultLang {
			continue
		}
		// Prefer language-specific domain if configured
		langHost := hostName
		if d, ok := config.LanguageDomains[lang.Id]; ok && d != "" {
			if httpRegex.MatchString(d) {
				langHost = extractDomain(d)
			} else {
				langHost = d
			}
		}
		href := GetLink(
			"content_item",
			config,
			langHost,
			lang.Id,
			true,
			"slug", contentItem.Slug,
			"id", contentItem.Id,
			"categories", contentItem.Categories,
			"full_url", true,
		)
		if href == "" {
			continue
		}
		if _, err := fmt.Fprintf(writer, `<xhtml:link rel="alternate" hreflang="%s" href="%s"/></xhtml:link>`, lang.Id, href); err != nil {
			return &pongo2.Error{Sender: "tag:sitemap_alternates", OrigError: err}
		}
	}
	return nil
}

func pongo2SitemapAlternates(_ *pongo2.Parser, _ *pongo2.Token, _ *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	return &tagSitemapAlternatesNode{}, nil
}
