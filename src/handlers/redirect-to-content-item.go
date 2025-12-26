package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"sersh.com/totaltube/frontend/types"
)

var RedirectToContentItem = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	config := r.Context().Value(types.ContextKeyConfig).(*types.Config)
	langId := r.Context().Value(types.ContextKeyLang).(string)
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	slug := r.URL.Query().Get("slug")
	if id <= 0 && slug == "" {
		Output404(w, r, "page not found")
		return
	}
	relatedRandomizeLast := 0
	if config.Related.Randomize != nil {
		relatedRandomizeLast = *config.Related.Randomize
	} else if internal.Config.Related.Randomize != nil {
		relatedRandomizeLast = *internal.Config.Related.Randomize
	}
	relatedParams := &api.RelatedParams{
		TitleTranslated:              config.Related.TitleTranslated,
		TitleTranslatedMinTermFreq:   config.Related.TitleTranslatedMinTermFreq,
		TitleTranslatedMaxQueryTerms: config.Related.TitleTranslatedMaxQueryTerms,
		TitleTranslatedBoost:         config.Related.TitleTranslatedBoost,
		Title:                        config.Related.Title,
		TitleMinTermFreq:             config.Related.TitleMinTermFreq,
		TitleMaxQueryTerms:           config.Related.TitleMaxQueryTerms,
		TitleBoost:                   config.Related.TitleBoost,
		Tags:                         config.Related.Tags,
		TagsMinTermFreq:              config.Related.TagsMinTermFreq,
		TagsMaxQueryTerms:            config.Related.TagsMaxQueryTerms,
		TagsBoost:                    config.Related.TagsBoost,
		RandomizeLast:                relatedRandomizeLast,
	}
	results, err := api.ContentItem(config, langId, slug, id, true, 0, 0, relatedParams)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			Output404(w, r, "content item not found")
			return
		}
		Output500(w, r, err)
		return
	}
	link := site.GetLink("content-item", config, hostName, langId, false, "slug", results.Slug, "id", results.Id, "categories", results.Categories)
	http.Redirect(w, r, link, http.StatusFound)
	if internal.Config.General.EnableAccessLog {
		log.Println(hostName, http.StatusFound, link)
	}
})
