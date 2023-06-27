package site

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v4"
	"github.com/sersh88/timeago"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

type PaginationItemType string

const (
	PaginationItemTypePage     = "page"
	paginationItemTypeEllipsis = "ellipsis"
	PaginationItemTypePrev     = "prev"
	PaginationItemTypeNext     = "next"
)

type PaginationItemState string

const (
	PaginationItemStateActive   = "active"
	PaginationItemStateDefault  = "default"
	PaginationItemStateDisabled = "disabled"
)

type PaginationItem struct {
	Type  string
	State string
	Page  int64
}
type IframeParsed struct {
	Src    string
	Width  int64
	Height int64
}

var iframeSrcRegex = regexp.MustCompile(`(?i)<\s*iframe[^>]*\ssrc\s*=\s*['"]?([^'" >]+)`)
var iframeWidthRegex = regexp.MustCompile(`(?i)<\s*iframe[^>]*\swidth\s*=\s*['"]?([^'" >]+)`)
var iframeHeightRegex = regexp.MustCompile(`(?i)<\s*iframe[^>]*\sheight\s*=\s*['"]?([^'" >]+)`)
var iframeHttpReplace = regexp.MustCompile(`(?i)^http://`)

func generateContext(name string, sitePath string, customContext pongo2.Context) pongo2.Context {
	refreshTranslations, _ := customContext["refreshTranslations"].(bool)
	var ctx = pongo2.Context{
		"flate":        helpers.Flate,
		"deflate":      helpers.Deflate,
		"gzip":         helpers.Gzip,
		"ungzip":       helpers.Ungzip,
		"zip":          helpers.Zip,
		"unzip":        helpers.Unzip,
		"base64":       helpers.Base64,
		"sha1":         helpers.Sha1Hash,
		"sha1_raw":     helpers.Sha1HashRaw,
		"md5":          helpers.Md5Hash,
		"md5_raw":      helpers.Md5HashRaw,
		"md4":          helpers.Md4Hash,
		"md4_raw":      helpers.Md4HashRaw,
		"sha256":       helpers.Sha256Hash,
		"sha256_raw":   helpers.Sha256HashRaw,
		"sha512":       helpers.Sha512Hash,
		"sha512_raw":   helpers.Sha512HashRaw,
		"time8601":     helpers.Time8601,
		"duration8601": helpers.Duration8601,
		"translate": func(text interface{}) interface{} {
			return deferredTranslate("en", customContext["lang"].(*types.Language).Id, text, "page-text", refreshTranslations)
		},
		"translate_title": func(text interface{}) interface{} {
			return deferredTranslate("en", customContext["lang"].(*types.Language).Id, text, "content-title", refreshTranslations)
		},
		"translate_description": func(text interface{}) interface{} {
			return deferredTranslate("en", customContext["lang"].(*types.Language).Id, text, "content-description", refreshTranslations)
		},
		"translate_query": func(text interface{}) interface{} {
			return deferredTranslate("en", customContext["lang"].(*types.Language).Id, text, "query", refreshTranslations)
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
		"len": func(items *pongo2.Value) int {
			return items.Len()
		},
		"link": func(route string, args ...interface{}) string {
			config, _ := customContext["config"].(*Config)
			pageTemplate, _ := customContext["page_template"].(string)
			lang, _ := customContext["lang"].(*types.Language)
			if route == "current" {
				if args == nil {
					args = make([]interface{}, 0)
				}
				route = customContext["route"].(string)
				currentParams := customContext["params"].(map[string]string)
				for k, v := range currentParams {
					args = append(args, k, v)
				}
				queryParams := url.Values(http.Header(customContext["canonical_query"].(url.Values)).Clone())
				for k, v := range queryParams {
					args = append(args, k, v)
				}
			}
			return GetLink(route, config, pageTemplate, lang.Id, args...)
		},
		"time_ago": func(t time.Time) string {
			langId := strings.ReplaceAll(customContext["lang"].(*types.Language).Id, "-", "_")
			return timeago.New(t).WithLocale(langId).Format()
		},
		"format_duration": func(duration int32) string {
			var d = time.Duration(duration)
			return fmt.Sprintf("%d:%d", int(d.Minutes()), int(d.Seconds()))
		},
		"pagination": func() (result []PaginationItem) {
			if pages, ok := customContext["pages"].(int64); ok && pages > 0 {
				var pagination = make([]PaginationItem, 0, pages+5)
				config := customContext["config"].(*Config)
				page := customContext["page"].(int64)
				var prevState = PaginationItemStateDefault
				if page == 1 {
					prevState = PaginationItemStateDisabled
				}
				pagination = append(pagination, PaginationItem{Type: PaginationItemTypePrev, State: prevState, Page: page - 1})
				var beforeCurrentPageLinks int64
				var beforeEllipsis bool
				var afterCurrentPageLinks int64
				var afterEllipsis bool
				if config.General.PaginationMaxRenderedLinks > 0 && config.General.PaginationMaxRenderedLinks < int(pages-2) {
					// amount of rendered links before current
					beforeCurrentPageLinks = int64(math.Min(float64(page-1), math.Round(float64(config.General.PaginationMaxRenderedLinks)/2)))
					if beforeCurrentPageLinks < page-1 {
						beforeEllipsis = true
					}
					afterCurrentPageLinks = int64(math.Min(
						float64(int64(config.General.PaginationMaxRenderedLinks)-beforeCurrentPageLinks-1),
						float64(pages-beforeCurrentPageLinks-1),
					))
					if afterCurrentPageLinks < pages-page {
						afterEllipsis = true
					}
				}
				for p := int64(1); p <= pages; p++ {
					if beforeCurrentPageLinks > 0 && p < page-beforeCurrentPageLinks {
						if p == 1 && beforeEllipsis {
							// show first page link and ellipsis
							pagination = append(pagination, PaginationItem{Type: PaginationItemTypePage, State: PaginationItemStateDefault, Page: p})
							if p < page-beforeCurrentPageLinks-1 {
								pagination = append(pagination, PaginationItem{Type: paginationItemTypeEllipsis})
							}
						}
						continue // do not render some items before current page
					}
					if afterCurrentPageLinks > 0 && p > page+afterCurrentPageLinks {
						if p == pages && afterEllipsis {
							// show ellipsis and last page link
							if p > page+afterCurrentPageLinks+1 {
								pagination = append(pagination, PaginationItem{Type: paginationItemTypeEllipsis})
							}
							pagination = append(pagination, PaginationItem{
								Type:  PaginationItemTypePage,
								State: PaginationItemStateDefault,
								Page:  p,
							})
						}
						continue // do not render some items after current page
					}
					var pageState = PaginationItemStateDefault
					if page == p {
						pageState = PaginationItemStateActive
					}
					pagination = append(pagination, PaginationItem{Type: PaginationItemTypePage, State: pageState, Page: p})
				}
				nextState := PaginationItemStateDefault
				if page == pages {
					nextState = PaginationItemStateDisabled
				}
				pagination = append(pagination, PaginationItem{Type: PaginationItemTypeNext, State: nextState, Page: page + 1})
				return pagination
			} else {
				return []PaginationItem{}
			}
		},
		"parse_iframe": func(iframeI interface{}) (parsed IframeParsed) {
			var iframe string
			if iframeP, ok := iframeI.(*string); ok {
				if iframeP != nil {
					iframe = *iframeP
				}
			} else if iframe, ok = iframeI.(string); !ok {
				log.Printf("Wrong iframe param type - %T in parse_iframe function", iframeI)
				return
			}
			matches := iframeSrcRegex.FindStringSubmatch(iframe)
			if matches == nil {
				log.Println("wrong iframe -", iframe, "in parse_iframe function")
				return
			}
			matches[1] = strings.ReplaceAll(matches[1], "https://", "http://")
			parsed.Src = iframeHttpReplace.ReplaceAllString(matches[1], "https://")
			matches = iframeWidthRegex.FindStringSubmatch(iframe)
			if matches == nil {
				log.Println("wrong iframe -", iframe, "in parse_iframe function")
				return
			}
			parsed.Width, _ = strconv.ParseInt(matches[1], 10, 64)
			matches = iframeHeightRegex.FindStringSubmatch(iframe)
			if matches == nil {
				log.Println("wrong iframe -", iframe, "in parse_iframe function")
				return
			}
			parsed.Height, _ = strconv.ParseInt(matches[1], 10, 64)
			return
		},
	}
	return ctx.Update(customContext)
}
