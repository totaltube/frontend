package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/middlewares"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

// ToplistData will handle requests to get most clickable thumbs for trading with other sites
var ToplistData = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	nocache, _ := strconv.ParseBool(r.URL.Query().Get(config.Params.Nocache))
	query := r.URL.Query().Get("query")
	lang := r.URL.Query().Get("lang")
	noParams := (query == "") && (lang == "")
	if lang == "" {
		lang = config.General.DefaultLanguage
	}
	lang = strings.ToLower(strings.Split(lang, "-")[0])
	if lang == "zh" {
		lang = "zh-Hans"
	}
	ip := r.Context().Value(types.ContextKeyIp).(string)
	groupId := internal.DetectCountryGroup(net.ParseIP(ip)).Id
	var additionalLanguages []string
	/*if noParams {
		additionalLanguages = lo.Map(internal.GetLanguages(config), func(language types.Language, _ int) string {
			return language.Id
		})
		additionalLanguages = lo.Filter(additionalLanguages, func(language string, _ int) bool {
			return language != config.General.DefaultLanguage
		})
	}*/
	cacheKey := fmt.Sprintf(`td:%s`, helpers.Md5Hash(fmt.Sprintf(`%s-%s-%s-%d-%v`, hostName, query, lang, groupId, noParams)))
	cacheTtl := time.Minute * 30

	var mapToplistRes = func(item *types.ContentResult, _ int) types.ToplistItem {
		description := ""
		if item.Description != nil {
			description = *item.Description
		}
		thumb := item.Thumb()
		hiresThumb := item.HiresThumb()
		if hiresThumb == thumb {
			hiresThumb = ""
		}
		category := "default"
		if len(item.Categories) > 0 {
			category = item.Categories[0].Slug
		}
		return types.ToplistItem{
			Title:                   item.Title,
			Description:             description,
			TitleTranslations:       item.TitleTranslations,
			DescriptionTranslations: item.DescriptionTranslations,
			Thumb:                   thumb,
			HiresThumb:              hiresThumb,
			ContentData: types.ToplistContentData{
				ContentId: item.Id,
				Url:       site.GetLink("content-item", config, hostName, lang, false, "full_url", true, "id", item.Id, "slug", item.Slug, "category", category),
			},
		}
	}
	result, err := db.GetCachedTimeout(cacheKey, cacheTtl, cacheTtl/2, func() (result []byte, err error) {
		var amount int64 = 50
		var toplistResults types.ToplistResults
		toplistResults.Items = make([]types.ToplistItem, 0, 50)
		toplistResults.Success = true
		if noParams {
			if links := buildToplistLanguageLinks(config, hostName); len(links) > 0 {
				toplistResults.LanguageLinks = links
			}
		}
		if query != "" {
			var queryResult json.RawMessage
			queryResult, err = api.ContentRaw(config, api.ContentParams{
				Amount:              amount,
				Lang:                lang,
				Sort:                api.SortPopular,
				SearchQuery:         query,
				GroupId:             groupId,
				Page:                1,
				AdditionalLanguages: additionalLanguages,
			})
			if err != nil {
				log.Println(err)
				return
			}
			contentResults := new(types.ContentResults)
			err = json.Unmarshal(queryResult, contentResults)
			if err != nil {
				log.Println(err)
				return
			}
			toplistResults.Items = append(toplistResults.Items, lo.Map(contentResults.Items, mapToplistRes)...)
			if len(toplistResults.Items) >= int(amount) {
				result, err = json.Marshal(toplistResults)
				return
			}
			amount = amount - int64(len(toplistResults.Items))
		}
		// all remaining items will be taken from popular
		var popularResult json.RawMessage
		popularResult, err = api.ContentRaw(config, api.ContentParams{
			Amount:              amount,
			Lang:                lang,
			Sort:                api.SortPopular,
			GroupId:             groupId,
			Page:                1,
			AdditionalLanguages: additionalLanguages,
		})
		if err != nil {
			log.Println(err)
			return
		}
		contentResults := new(types.ContentResults)
		err = json.Unmarshal(popularResult, contentResults)
		if err != nil {
			log.Println(err)
			return
		}
		toplistResults.Items = append(toplistResults.Items, lo.Map(contentResults.Items, mapToplistRes)...)
		result, err = json.Marshal(toplistResults)
		return
	}, nocache)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, M{"success": false, "error": err.Error()})
		return
	}
	if middlewares.HeadersSent(w) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Robots-Tag", "noindex")
	_, _ = w.Write(result)
})

func buildToplistLanguageLinks(config *types.Config, fallbackHost string) map[string]string {
	route := config.General.RouteForToplistLanguageLinks
	if route == "" {
		if config.Routes.TopContent == "/" {
			route = "top-content"
		} else if config.Routes.TopCategories == "/" {
			route = "top-categories"
		} else if config.Routes.Popular == "/" {
			route = "popular"
		} else if config.Routes.New == "/" {
			route = "new"
		}
	}
	if config == nil || !config.General.MultiLanguage || !config.General.IncludeToplistLanguageLinks || route == "" {
		return nil
	}
	languages := internal.GetLanguages(config)
	if len(languages) == 0 {
		return nil
	}
	defaultTarget, _ := internal.GetDefaultLanguageDomainTarget(config)
	languageLinks := make(map[string]string, len(languages))
	for _, language := range languages {
		if language.Id == "" {
			continue
		}
		changeLangLink := true
		if config.General.NoRedirectDefaultLanguage && language.Id == config.General.DefaultLanguage {
			changeLangLink = false
		}
		hostForLink := strings.TrimSpace(fallbackHost)
		if domainValue := strings.TrimSpace(config.LanguageDomains[language.Id]); domainValue != "" {
			if target, ok := internal.ParseLanguageDomainTarget(domainValue); ok {
				hostForLink = target.Host
			} else {
				hostForLink = internal.NormalizeHost(domainValue)
			}
		} else if defaultTarget != nil && defaultTarget.Host != "" {
			hostForLink = defaultTarget.Host
		} else if hostForLink == "" {
			hostForLink = config.Hostname
		}
		if hostForLink == "" {
			continue
		}
		args := make([]any, 0, 6)
		args = append(args, "full_url", true)
		link := site.GetLink(route, config, hostForLink, language.Id, changeLangLink, args...)
		if link == "" {
			continue
		}
		languageLinks[language.Id] = link
		languageIdShort := strings.Split(language.Id, "-")[0]
		if languageIdShort != language.Id {
			languageLinks[languageIdShort] = link
		}
	}
	if len(languageLinks) == 0 {
		return nil
	}
	return languageLinks
}
