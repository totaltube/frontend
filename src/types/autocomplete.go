package types

type AutocompleteItem struct {
	Suggest string `json:"suggest"`
	Lang    string `json:"lang"`
	Type    string `json:"type,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Id      int64  `json:"id,omitempty"`
	Total   int64  `json:"total"`
}

type AutocompleteResults struct {
	Items []AutocompleteItem `json:"items"`
}
