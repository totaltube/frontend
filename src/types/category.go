package types

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/samber/lo"
)

type CtrsStatus string

const (
	CtrsStatusNormal  = "normal"
	CtrsStatusCasting = "casting"
)

type CategoryResult struct {
	Id                 int32               `json:"id"`
	Slug               string              `json:"slug"`
	Title              string              `json:"title"`
	TitleTranslated    bool                `json:"title_translated,omitempty"`
	OriginalTitle      string              `json:"original_title"`
	Description        *string             `json:"description"`
	Tags               []string            `json:"tags"`
	Dated              time.Time           `json:"dated"`
	CreatedAt          time.Time           `json:"created_at"`
	AliasCategoryId    *int32              `json:"alias_category_id,omitempty"`
	ThumbRetina        bool                `json:"thumb_retina"`
	ThumbWidth         int32               `json:"thumb_width"`
	ThumbHeight        int32               `json:"thumb_height"`
	ThumbsAmount       int32               `json:"thumbs_amount"`
	ThumbsServer       string              `json:"thumbs_server"`
	ThumbsPath         string              `json:"thumbs_path"`
	ThumbFormat        string              `json:"thumb_format"`
	ThumbType          string              `json:"thumb_type"`
	ThumbFormats       []ThumbFormat       `json:"thumb_formats"`
	BestThumb          *int16              `json:"best_thumb,omitempty"`
	RotationStatus     *CtrsStatus         `json:"rotation_status,omitempty"`
	Total              int32               `json:"total"`
	Ctr                *float32            `json:"ctr,omitempty"`
	Views              int32               `json:"views"`
	CustomData         CustomData         `json:"custom_data"`
	CustomTranslations CustomTranslations `json:"custom_translations"`
	selectedThumb      *int
}
type CategoryResults struct {
	Total int               `json:"total"`
	From  int               `json:"from"`
	To    int               `json:"to"`
	Page  int               `json:"page"`
	Pages int               `json:"pages"` // total pages
	Items []*CategoryResult `json:"items"`
}

func (c CategoryResult) GetThumbFormat(thumbFormatName ...string) (res ThumbFormat) {
	if len(c.ThumbFormats) == 0 {
		return
	}
	res = c.ThumbFormats[0]
	if len(thumbFormatName) > 0 {
		for _, name := range thumbFormatName {
			name := name
			if f, ok := lo.Find(c.ThumbFormats, func(tf ThumbFormat) bool { return tf.Name == name }); ok {
				res = f
				return
			}
		}
	}
	return
}

func (c CategoryResult) ThumbTemplate(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + format.Name + ".%d." + format.Type
}

func (c *CategoryResult) Thumb(thumbFormatName ...string) string {
	return fmt.Sprintf(c.ThumbTemplate(thumbFormatName...), c.SelectedThumb(thumbFormatName...))
}

func (c *CategoryResult) HiresThumb(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	if format.Retina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(thumbFormatName...), "."+format.Type)+"@2x."+format.Type, c.SelectedThumb(thumbFormatName...))
	} else {
		return c.Thumb(thumbFormatName...)
	}
}

func (c *CategoryResult) SelectedThumb(thumbFormatName ...string) int {
	if c.selectedThumb != nil {
		return *c.selectedThumb
	}
	if c.BestThumb != nil {
		idx := int(*c.BestThumb)
		c.selectedThumb = &idx
	} else {
		format := c.GetThumbFormat(thumbFormatName...)
		var idx int
		if format.Amount > 0 {
			idx = rand.Intn(int(format.Amount))
		}
		c.selectedThumb = &idx
	}
	return *c.selectedThumb
}

func (c CategoryResult) HasCustomField(name string) bool {
	if c.CustomData == nil {
		return false
	}
	_, ok := c.CustomData[name]
	return ok
}

func (c CategoryResult) CustomField(name string) interface{} {
	if c.CustomData == nil {
		return nil
	}
	if data, ok := c.CustomData[name]; ok {
		return data
	}
	return nil
}

func (c CategoryResult) HasCustomTranslation(key string) bool {
	if c.CustomTranslations == nil {
		return false
	}
	_, ok := c.CustomTranslations[key]
	return ok
}

func (c CategoryResult) CustomTranslation(key string) (translation string) {
	if c.CustomTranslations != nil {
		translation, _ = c.CustomTranslations[key]
	}
	if translation == "" && c.CustomData != nil {
		// if we don't have translation for current language, maybe we will find original text in CustomData
		if customData, ok := c.CustomData[key]; ok {
			translation, _ = customData.(string)
		}
	}
	return
}
