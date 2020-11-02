package types

type RelatedItem struct {
	Message string      `json:"message"`
	Type    ContentType `json:"type,omitempty"`
	Id      int64       `json:"id,omitempty"`
	Slug    string      `json:"slug,omitempty"`
}
