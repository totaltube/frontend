package types

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type ModelResult struct {
	Id              int32     `json:"id"`
	Slug            string    `json:"slug"`
	Title           string    `json:"title"`
	TitleTranslated bool      `json:"title_translated,omitempty"`
	Description     *string   `json:"description"`
	Dated           time.Time `json:"dated"`
	CreatedAt       time.Time `json:"created_at"`
	ThumbRetina     bool      `json:"thumb_retina"`
	ThumbWidth      int32     `json:"thumb_width"`
	ThumbHeight     int32     `json:"thumb_height"`
	ThumbsAmount    int32     `json:"thumbs_amount"`
	ThumbsServer    string    `json:"thumbs_server"`
	ThumbsPath      string    `json:"thumbs_path"`
	ThumbFormat     string    `json:"thumb_format"`
	ThumbType       string    `json:"thumb_type"`
	BestThumb       *int16    `json:"best_thumb,omitempty"`
	Total           int32     `json:"total"`
	Views           int32     `json:"views"`
	selectedThumb   *int
}
type ModelResults struct {
	Total int            `json:"total"` // Всего категорий
	From  int            `json:"from"`  // с какого элемента показываются результаты
	To    int            `json:"to"`    // до какого элемента показываются результаты
	Page  int            `json:"page"`
	Pages int            `json:"pages"`
	Items []*ModelResult `json:"items"` // выбранные результаты
}

func (c ModelResult) ThumbTemplate() string {
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + c.ThumbFormat + ".%d."+c.ThumbType
}

func (c *ModelResult) Thumb() string {
	return fmt.Sprintf(c.ThumbTemplate(), c.SelectedThumb())
}

func (c *ModelResult) HiresThumb() string {
	if c.ThumbRetina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(), "."+c.ThumbType)+"@2x."+c.ThumbType, c.SelectedThumb())
	} else {
		return c.Thumb()
	}
}

func (c *ModelResult) SelectedThumb() int {
	if c.selectedThumb != nil {
		return *c.selectedThumb
	}
	if c.BestThumb != nil {
		idx := int(*c.BestThumb)
		c.selectedThumb = &idx
	} else {
		idx := rand.Intn(int(c.ThumbsAmount))
		c.selectedThumb = &idx
	}
	return *c.selectedThumb
}
