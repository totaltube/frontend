package types

import "time"

type CtrsStatus string

const (
	CtrsStatusNormal  = "normal"
	CtrsStatusCasting = "casting"
)

type CategoryResult struct {
	Id              int32       `json:"id"`
	Slug            string      `json:"slug"`
	Title           string      `json:"title"`
	TitleTranslated bool        `json:"title_translated,omitempty"`
	Description     *string     `json:"description"`
	Tags            []string    `json:"tags"`
	Dated           time.Time   `json:"dated"`
	CreatedAt       time.Time   `json:"created_at"`
	AliasCategoryId *int32      `json:"alias_category_id,omitempty"`
	ThumbRetina     bool        `json:"thumb_retina"`
	ThumbWidth      int32       `json:"thumb_width"`
	ThumbHeight     int32       `json:"thumb_height"`
	ThumbsAmount    int32       `json:"thumbs_amount"`
	ThumbsServer    string      `json:"thumbs_server"`
	ThumbsPath      string      `json:"thumbs_path"`
	BestThumb       *int16      `json:"best_thumb,omitempty"`
	RotationStatus  *CtrsStatus `json:"rotation_status,omitempty"`
	Total           int32       `json:"total"`
	Ctr             *float32    `json:"ctr,omitempty"`
	Views           int32       `json:"views"`
}
type CategoryResults struct {
	Total int              `json:"total"` // Всего категорий
	From  int              `json:"from"`  // с какого элемента показываются результаты
	To    int              `json:"to"`    // до какого элемента показываются результаты
	Items []CategoryResult `json:"items"` // выбранные результаты
}
