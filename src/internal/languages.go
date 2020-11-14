package internal

import (
	"golang.org/x/text/language"
	"sersh.com/totaltube/frontend/types"
	"strings"
)

var languages []types.Language
var languagesMap map[string]*types.Language
var tagsMap map[string]*types.Language
var languageTags []language.Tag
var matcher language.Matcher

func InitLanguages(Languages []types.Language) {
	languages = Languages
	enOk := false
	for k := range languages {
		if languages[k].Name == "en" {
			enOk = true
		}
	}
	if !enOk {
		languages = append(languages, types.Language{
			Name:      "en",
			Locale:    "en_US",
			Native:    "English",
			Direction: types.LanguageDirectionLtr,
			Country:   "us",
		})
	}
	languagesMap = make(map[string]*types.Language)
	tagsMap = make(map[string]*types.Language)
	languageTags = make([]language.Tag, 0, len(languages))
	for k := range languages {
		languages[k].Tag = language.MustParse(strings.Replace(languages[k].Locale, "_", "-", -1))
		languagesMap[languages[k].Name] = &languages[k]
		tagsMap[languages[k].Tag.String()] = &languages[k]
		languageTags = append(languageTags, languages[k].Tag)
	}
	matcher = language.NewMatcher(languageTags)
}

func GetLanguages() []types.Language {

	return languages
}

func GetLanguage(lang string) *types.Language {
	if l, ok := languagesMap[lang]; ok {
		return l
	}
	return nil
}

func DetectLanguage(langCookie, acceptLanguageHeader string) *types.Language {
	tag, _ := language.MatchStrings(matcher, langCookie, acceptLanguageHeader)
	if l, ok := tagsMap[tag.String()]; ok {
		return l
	}
	return nil
}
