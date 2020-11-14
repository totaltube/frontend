package types

import "golang.org/x/text/language"

type LanguageDirection string

const (
	LanguageDirectionLtr LanguageDirection = "ltr"
	LanguageDirectionRtl LanguageDirection = "rtl"
)

type Language struct {
	Name      string
	Locale    string
	Native    string
	Direction LanguageDirection
	Country   string
	Tag       language.Tag
}
