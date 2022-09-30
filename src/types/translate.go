package types

type TranslateParams struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
	Type string `json:"type"`
}

var TranslationTypes = []string{"query", "content-title", "content-description", "taxonomy-title", "taxonomy-description", "page-text"}