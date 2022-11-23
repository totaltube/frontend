package types

import "golang.org/x/text/language"

type LanguageDirection string

const (
	LanguageDirectionLtr LanguageDirection = "ltr"
	LanguageDirectionRtl LanguageDirection = "rtl"
)

type Language struct {
	Id        string            // language ID like "en", "de", "it"
	Name      string            // language name like English, German, Italian
	Locale    string            // language locale like "en_US", "de_DE", "it_IT"
	Native    string            // native language name like English, Deutsch, Italiano
	Direction LanguageDirection // language direction: "ltr" (left to right) or "rtl" (right to left)
	Country   string            // country code associated with language like us, de, it
	Tag       language.Tag      // represents a BCP 47 language tag. It is used to specify an instance of a specific language or locale.
}
