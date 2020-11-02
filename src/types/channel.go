package types

import "time"

type ChannelResult struct {
	Id              int32     `json:"id"`
	Slug            string    `json:"slug"`
	Title           string    `json:"title"`
	TitleTranslated bool      `json:"title_translated,omitempty"`
	Description     *string   `json:"description"`
	Tags            []string  `json:"tags"`
	Dated           time.Time `json:"dated"`
	CreatedAt       time.Time `json:"created_at"`
	ThumbRetina     bool      `json:"thumb_retina"`
	ThumbWidth      int32     `json:"thumb_width"`
	ThumbHeight     int32     `json:"thumb_height"`
	ThumbsAmount    int32     `json:"thumbs_amount"`
	ThumbsServer    string    `json:"thumbs_server"`
	ThumbsPath      string    `json:"thumbs_path"`
	BestThumb       *int16    `json:"best_thumb,omitempty"`
	Total           int32     `json:"total"`
	Views           int32     `json:"views" db:"views"`
}

type ChannelResults struct {
	Total int             `json:"total"` // Всего категорий
	From  int             `json:"from"`  // с какого элемента показываются результаты
	To    int             `json:"to"`    // до какого элемента показываются результаты
	Items []ChannelResult `json:"items"` // выбранные результаты
}
