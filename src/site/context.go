package site

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/samber/lo"
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
	var mu sync.Mutex
	refreshTranslations, _ := customContext["refreshTranslations"].(bool)
	translateFunc := func(text any, params ...any) any {
		langFrom := "en"
		langTo := customContext["lang"].(*types.Language).Id
		Type := "page-text"
		currentIndex := 0
		for currentIndex+1 < len(params) {
			if param, ok := params[currentIndex].(string); ok {
				switch param {
				case "from":
					if langFromParam, ok := params[currentIndex+1].(*types.Language); ok {
						langFrom = langFromParam.Id
					} else if langFromParam, ok := params[currentIndex+1].(string); ok {
						langFrom = langFromParam
					}
				case "to":
					if langToParam, ok := params[currentIndex+1].(*types.Language); ok {
						langTo = langToParam.Id
					} else if langToParam, ok := params[currentIndex+1].(string); ok {
						langTo = langToParam
					}
				case "type":
					if TypeParam, ok := params[currentIndex+1].(string); ok && lo.Contains(types.TranslationTypes, TypeParam) {
						Type = TypeParam
					}
				}
				currentIndex += 2
			}
		}
		return deferredTranslate(langFrom, langTo, text, Type, refreshTranslations, customContext["config"].(*types.Config))
	}
	var ctx = pongo2.Context{
		"flate":          helpers.Flate,
		"deflate":        helpers.Deflate,
		"bytes":          helpers.Bytes,
		"gzip":           helpers.Gzip,
		"ungzip":         helpers.Ungzip,
		"zip":            helpers.Zip,
		"unzip":          helpers.Unzip,
		"base64":         helpers.Base64,
		"base64_url":     helpers.Base64Url,
		"base64_raw_url": helpers.Base64RawUrl,
		"htmlentities":   helpers.HtmlEntitiesAll,
		"sha1":           helpers.Sha1Hash,
		"sha1_raw":       helpers.Sha1HashRaw,
		"md5":            helpers.Md5Hash,
		"md5_raw":        helpers.Md5HashRaw,
		"md4":            helpers.Md4Hash,
		"md4_raw":        helpers.Md4HashRaw,
		"sha256":         helpers.Sha256Hash,
		"sha256_raw":     helpers.Sha256HashRaw,
		"sha512":         helpers.Sha512Hash,
		"sha512_raw":     helpers.Sha512HashRaw,
		"time8601":       helpers.Time8601,
		"duration8601":   helpers.Duration8601,
		"slugify":        helpers.Slugify,
		"translate":      translateFunc,
		"translate_title": func(text any, params ...any) any {
			return translateFunc(text, append(params, "type", "content-title")...)
		},
		"translate_description": func(text any, params ...any) any {
			return translateFunc(text, append(params, "type", "content-description")...)
		},
		"translate_query": func(text any, params ...any) any {
			return translateFunc(text, append(params, "type", "query")...)
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
			mu.Lock()
			defer mu.Unlock()
			config, _ := customContext["config"].(*types.Config)
			lang, _ := customContext["lang"].(*types.Language)
			hostName, _ := customContext["host"].(string)
			langId := lang.Id
			var changeLangLink bool
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
			finalArgs := make([]interface{}, 0, len(args))
			for i := 0; i < len(args); i += 2 {
				if key, ok := args[i].(string); ok {
					if key == "lang" && len(args) > i+1 {
						if val, ok := args[i+1].(string); ok {
							if lang.Id != val {
								changeLangLink = true
								langId = val
							}
							continue
						}
					}
				}
				finalArgs = append(finalArgs, args[i], args[i+1])
			}
			return GetLink(route, config, hostName, langId, changeLangLink, finalArgs...)
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
				config := customContext["config"].(*types.Config)
				page := customContext["page"].(int64)
				if page > pages {
					return []PaginationItem{}
				}
				var prevState = PaginationItemStateDefault
				if page == 1 {
					prevState = PaginationItemStateDisabled
				}
				pagination = append(pagination, PaginationItem{Type: PaginationItemTypePrev, State: prevState, Page: page - 1})

				maxLinks := config.General.PaginationMaxRenderedLinks
				// If maxLinks is not set or is large enough, show all pages
				if maxLinks <= 0 || maxLinks >= int(pages-2) {
					for p := int64(1); p <= pages; p++ {
						var pageState = PaginationItemStateDefault
						if page == p {
							pageState = PaginationItemStateActive
						}
						pagination = append(pagination, PaginationItem{Type: PaginationItemTypePage, State: pageState, Page: p})
					}
				} else {
					// Calculate window around current page
					// We want to show maxLinks pages total, with current page in the middle when possible
					maxLinks64 := int64(maxLinks)

					// Calculate how many pages to show before and after current page
					// Try to balance: if we have odd number, give one extra to the side with more pages
					windowHalf := (maxLinks64 - 1) / 2
					windowStart := page - windowHalf
					windowEnd := page + windowHalf

					// Adjust window if it goes beyond boundaries
					if windowStart < 1 {
						windowStart = 1
						windowEnd = min(maxLinks64, pages)
					}
					if windowEnd > pages {
						windowEnd = pages
						windowStart = max(1, pages-maxLinks64+1)
					}

					// Determine if we need ellipsis
					needBeforeEllipsis := windowStart > 2
					needAfterEllipsis := windowEnd < pages-1

					// Show first page if needed
					if windowStart > 1 {
						pagination = append(pagination, PaginationItem{Type: PaginationItemTypePage, State: PaginationItemStateDefault, Page: 1})
						if needBeforeEllipsis {
							pagination = append(pagination, PaginationItem{Type: paginationItemTypeEllipsis})
						}
					}

					// Show pages in window
					for p := windowStart; p <= windowEnd; p++ {
						var pageState = PaginationItemStateDefault
						if page == p {
							pageState = PaginationItemStateActive
						}
						pagination = append(pagination, PaginationItem{Type: PaginationItemTypePage, State: pageState, Page: p})
					}

					// Show ellipsis and last page if needed
					if windowEnd < pages {
						if needAfterEllipsis {
							pagination = append(pagination, PaginationItem{Type: paginationItemTypeEllipsis})
						}
						pagination = append(pagination, PaginationItem{
							Type:  PaginationItemTypePage,
							State: PaginationItemStateDefault,
							Page:  pages,
						})
					}
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
