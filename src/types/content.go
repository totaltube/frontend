package types

import (
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"math/rand"
	"strings"
	"time"
)

type ContentType string

const (
	ContentTypeVideoEmbed ContentType = "video-embed"
	ContentTypeVideoLink  ContentType = "video-link"
	ContentTypeVideo      ContentType = "video"
	ContentTypeGallery    ContentType = "gallery"
	ContentTypeLink       ContentType = "link"
	ContentTypeContent    ContentType = "content" // Обобщенно все виды контента, а не таксономий

	ContentTypeCategory  ContentType = "category"
	ContentTypeChannel   ContentType = "channel"
	ContentTypeModel     ContentType = "model"
	ContentTypeUniversal ContentType = "universal"
)

type TaxonomyResult struct {
	Id    int32  `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}
type TaxonomyResults []TaxonomyResult

type ContentResultUser struct {
	Id    int32  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}
type ContentItemResult struct {
	Id              int64             `json:"id"`
	Slug            string            `json:"slug"`
	Title           string            `json:"title"`
	TitleTranslated bool              `json:"title_translated,omitempty"`
	Description     *string           `json:"description,omitempty"`
	Channel         *TaxonomyResult   `json:"channel,omitempty"`
	Content         *string           `json:"content,omitempty"`
	Link            *string           `json:"link,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	Dated           time.Time         `json:"dated"`
	Duration        int32             `json:"duration"`
	Tags            []string          `json:"tags"`
	Keywords        []string          `json:"keywords,omitempty"`
	ThumbsAmount    int32             `json:"thumbs_amount"`
	ThumbsWidth     int32             `json:"thumb_width"`
	ThumbsHeight    int32             `json:"thumb_height"`
	ThumbsServer    string            `json:"thumbs_server"` // урл сервера, на котором тумбы
	ThumbsPath      string            `json:"thumbs_path"`   // шаблон пути к тумбам
	ThumbRetina     bool              `json:"thumb_retina"`  // индикатор, что есть версия @2x
	ThumbFormat     string            `json:"thumb_format"`
	BestThumb       *int16            `json:"best_thumb,omitempty"` // номер лучшей тумбы, нумерация с 0
	Type            ContentType       `json:"type"`
	Priority        int16             `json:"priority,omitempty"`
	User            ContentResultUser `json:"user"`
	Categories      TaxonomyResults   `json:"categories,omitempty"`
	Models          TaxonomyResults   `json:"models,omitempty"`
	Views           int32             `json:"views"`
	Related         []ContentResult   `json:"related,omitempty"` // похожие на этот контент
	selectedThumb   *int
}

type ContentResult struct {
	Id              int64             `json:"id"`
	Slug            string            `json:"slug"`
	Title           string            `json:"title"`
	TitleTranslated bool              `json:"title_translated,omitempty"`
	Description     *string           `json:"description,omitempty"`
	Channel         *TaxonomyResult   `json:"channel,omitempty"`
	Content         *string           `json:"content,omitempty"`
	Link            *string           `json:"link,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	Dated           time.Time         `json:"dated"`
	Duration        int32             `json:"duration"`
	Tags            []string          `json:"tags"`
	Keywords        []string          `json:"keywords,omitempty"`
	ThumbsAmount    int32             `json:"thumbs_amount"`
	ThumbsWidth     int32             `json:"thumb_width"`
	ThumbsHeight    int32             `json:"thumb_height"`
	ThumbsServer    string            `json:"thumbs_server"` // урл сервера, на котором тумбы
	ThumbsPath      string            `json:"thumbs_path"`   // шаблон пути к тумбам
	ThumbRetina     bool              `json:"thumb_retina"`  // индикатор, что есть версия @2x
	ThumbFormat     string            `json:"thumb_format"`
	BestThumb       *int16            `json:"best_thumb,omitempty"` // Лучшая тумба
	Type            ContentType       `json:"type"`
	Priority        int16             `json:"priority,omitempty"`
	User            ContentResultUser `json:"user"`
	Categories      TaxonomyResults   `json:"categories,omitempty"`
	Models          TaxonomyResults   `json:"models,omitempty"`
	RotationStatus  *CtrsStatus       `json:"rotation_status,omitempty"`
	Ctr             *float32          `json:"ctr,omitempty"`
	Views           int32             `json:"views"`
	selectedThumb   *int
}

type ContentResults struct {
	Total   int64            `json:"total"` // Всего контента
	From    int              `json:"from"`  // с какого элемента показываются результаты
	To      int              `json:"to"`    // до какого элемента показываются результаты
	Page    int              `json:"page"`
	Pages   int              `json:"pages"`
	Items   []*ContentResult `json:"items"`             // выбранные результаты
	Title   string           `json:"title,omitempty"`   // заголовок результата
	Related []RelatedItem    `json:"related,omitempty"` // Похожие на текущий запрос запросы, категории, модели и тд.
}

func (tr *TaxonomyResults) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		t := make(TaxonomyResults, 0)
		*tr = t
		return nil
	}
	return json.Unmarshal(b, tr)
}

func (tr TaxonomyResults) Value() (driver.Value, error) {
	return json.Marshal(tr)
}

func (c *ContentResultUser) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, c)
}

func (c ContentResultUser) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c ContentItemResult) ThumbTemplate() string {
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + c.ThumbFormat + ".%d.jpg"
}

func (c *ContentItemResult) Thumb() string {
	return fmt.Sprintf(c.ThumbTemplate(), c.SelectedThumb())
}

func (c *ContentItemResult) HiresThumb() string {
	if c.ThumbRetina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(), ".jpg")+"@2x.jpg", c.SelectedThumb())
	} else {
		return c.Thumb()
	}
}

func (c *ContentItemResult) SelectedThumb() int {
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

func (c ContentResult) ThumbTemplate() string {
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + c.ThumbFormat + ".%d.jpg"
}

func (c *ContentResult) Thumb() string {
	return fmt.Sprintf(c.ThumbTemplate(), c.SelectedThumb())
}

func (c *ContentResult) HiresThumb() string {
	if c.ThumbRetina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(), ".jpg")+"@2x.jpg", c.SelectedThumb())
	} else {
		return c.Thumb()
	}
}

func (c *ContentResult) SelectedThumb() int {
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
