package types

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/samber/lo"
)

type ChannelResult struct {
	Id              int32         `json:"id"`
	Slug            string        `json:"slug"`
	Title           string        `json:"title"`
	TitleTranslated bool          `json:"title_translated,omitempty"`
	Description     *string       `json:"description"`
	Tags            []string      `json:"tags"`
	Dated           time.Time     `json:"dated"`
	CreatedAt       time.Time     `json:"created_at"`
	ThumbRetina     bool          `json:"thumb_retina"`
	ThumbWidth      int32         `json:"thumb_width"`
	ThumbHeight     int32         `json:"thumb_height"`
	ThumbsAmount    int32         `json:"thumbs_amount"`
	ThumbsServer    string        `json:"thumbs_server"`
	ThumbsPath      string        `json:"thumbs_path"`
	ThumbFormat     string        `json:"thumb_format"`
	ThumbType       string        `json:"thumb_type"`
	ThumbFormats    []ThumbFormat `json:"thumb_formats"`
	BestThumb       *int16        `json:"best_thumb,omitempty"`
	Total           int32         `json:"total"`
	Views           int32         `json:"views" db:"views"`
	selectedThumb   *int
}

type ChannelResults struct {
	Total int              `json:"total"`
	From  int              `json:"from"`
	To    int              `json:"to"`
	Page  int              `json:"page"`
	Pages int              `json:"pages"`
	Items []*ChannelResult `json:"items"`
}

func (c ChannelResult) GetThumbFormat(thumbFormatName ...string) (res ThumbFormat) {
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

func (c ChannelResult) ThumbTemplate(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + format.Name + ".%d." + format.Type
}

func (c *ChannelResult) Thumb(thumbFormatName ...string) string {
	return fmt.Sprintf(c.ThumbTemplate(thumbFormatName...), c.SelectedThumb(thumbFormatName...))
}

func (c *ChannelResult) HiresThumb(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	if format.Retina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(thumbFormatName...), "."+format.Type)+"@2x."+format.Type, c.SelectedThumb(thumbFormatName...))
	} else {
		return c.Thumb(thumbFormatName...)
	}
}

func (c *ChannelResult) SelectedThumb(thumbFormatName ...string) int {
	if c.selectedThumb != nil {
		return *c.selectedThumb
	}
	if c.BestThumb != nil {
		idx := int(*c.BestThumb)
		c.selectedThumb = &idx
	} else {
		format := c.GetThumbFormat(thumbFormatName...)
		idx := rand.Intn(int(format.Amount))
		c.selectedThumb = &idx
	}
	return *c.selectedThumb
}
