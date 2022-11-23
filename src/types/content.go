package types

import (
	"database/sql/driver"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/segmentio/encoding/json"
)

const (
	ContentTypeVideoEmbed = "video-embed"
	ContentTypeVideoLink  = "video-link"
	ContentTypeVideo      = "video"
	ContentTypeGallery    = "gallery"
	ContentTypeLink       = "link"
	ContentTypeContent    = "content" // All types of content, not taxonomies
	ContentTypeCategory   = "category"
	ContentTypeChannel    = "channel"
	ContentTypeModel      = "model"
	ContentTypeUniversal  = "universal"
)

type Size struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

func (s Size) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%dx%d", s.Width, s.Height))
}
func (s *Size) UnmarshalJSON(b []byte) error {
	var ss string
	err := json.Unmarshal(b, &ss)
	if err != nil {
		return err
	}
	sizes := strings.Split(ss, "x")
	if len(sizes) != 2 {
		return errors.New("wrong size")
	}
	*s = Size{}
	s.Width, err = strconv.ParseInt(sizes[0], 10, 16)
	if err != nil {
		return err
	}
	s.Height, err = strconv.ParseInt(sizes[1], 10, 16)
	if err != nil {
		return err
	}
	return nil
}

type ContentGalleryInfo struct {
	Items        []Size `json:"items"`
	PreviewItems []Size `json:"preview_items"`
	Type         string `json:"type"`
	Name         string `json:"name"`
}
type ContentVideoInfo struct {
	Name           string
	Type           string  `json:"type"`
	Size           Size    `json:"size"`
	VideoBitrate   int32   `json:"video_bitrate"`
	AudioBitrate   int32   `json:"audio_bitrate"`
	PosterType     string  `json:"poster_type,omitempty"`
	TimelineType   string  `json:"timeline_type,omitempty"`
	TimelineSize   Size    `json:"timeline_size"`
	TimelineFrames int32   `json:"timeline_frames"`
	Duration       float64 `json:"duration"`
}

type ContentDuration int32
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
	Id              int64                          `json:"id"`
	Slug            string                         `json:"slug"`
	Title           string                         `json:"title"`
	TitleTranslated bool                           `json:"title_translated,omitempty"`
	Description     *string                        `json:"description,omitempty"`
	Channel         *TaxonomyResult                `json:"channel,omitempty"`
	Content         *string                        `json:"content,omitempty"`
	Link            *string                        `json:"link,omitempty"`
	CreatedAt       time.Time                      `json:"created_at"`
	Dated           time.Time                      `json:"dated"`
	Duration        ContentDuration                `json:"duration"`
	Tags            []string                       `json:"tags"`
	Keywords        []string                       `json:"keywords,omitempty"`
	VideoServer     string                         `json:"video_server,omitempty"`
	GalleryServer   string                         `json:"gallery_server,omitempty"`
	VideoPath       string                         `json:"video_path,omitempty"`
	GalleryPath     string                         `json:"gallery_path,omitempty"`
	GalleryItems    *map[string]ContentGalleryInfo `json:"gallery_items,omitempty"`
	VideoSizes      *map[string]ContentVideoInfo   `json:"video_sizes,omitempty"`
	ThumbsAmount    int32                          `json:"thumbs_amount"`
	ThumbsWidth     int32                          `json:"thumb_width"`
	ThumbsHeight    int32                          `json:"thumb_height"`
	ThumbsServer    string                         `json:"thumbs_server"` // thumb server url
	ThumbsPath      string                         `json:"thumbs_path"`   // path to thumbs on thumb server
	ThumbRetina     bool                           `json:"thumb_retina"`  // there is @2x version of thumb
	ThumbFormat     string                         `json:"thumb_format"`
	ThumbType       string                         `json:"thumb_type"`           // image type ( jpg, webp, png )
	BestThumb       *int16                         `json:"best_thumb,omitempty"` // best thumb indexed from 0
	Type            string                         `json:"type"`
	Priority        int16                          `json:"priority,omitempty"`
	User            ContentResultUser              `json:"user"`
	Categories      TaxonomyResults                `json:"categories,omitempty"`
	Models          TaxonomyResults                `json:"models,omitempty"`
	Views           int32                          `json:"views"`
	Related         []*ContentResult               `json:"related,omitempty"` // similar content
	selectedThumb   *int
}

type ContentResult struct {
	Id              int64                          `json:"id"`
	Slug            string                         `json:"slug"`
	Title           string                         `json:"title"`
	TitleTranslated bool                           `json:"title_translated,omitempty"`
	Description     *string                        `json:"description,omitempty"`
	Channel         *TaxonomyResult                `json:"channel,omitempty"`
	Content         *string                        `json:"content,omitempty"`
	Link            *string                        `json:"link,omitempty"`
	CreatedAt       time.Time                      `json:"created_at"`
	Dated           time.Time                      `json:"dated"`
	Duration        ContentDuration                `json:"duration"`
	Tags            []string                       `json:"tags"`
	Keywords        []string                       `json:"keywords,omitempty"`
	GalleryServer   string                         `json:"gallery_server"`
	VideoServer     string                         `json:"video_server"`
	GalleryPath     string                         `json:"gallery_path"`
	VideoPath       string                         `json:"video_path"`
	GalleryItems    *map[string]ContentGalleryInfo `json:"gallery_items,omitempty"`
	VideoSizes      *map[string]ContentVideoInfo   `json:"video_sizes,omitempty"`
	ThumbsAmount    int32                          `json:"thumbs_amount"`
	ThumbsWidth     int32                          `json:"thumb_width"`
	ThumbsHeight    int32                          `json:"thumb_height"`
	ThumbsServer    string                         `json:"thumbs_server"` // thumb server url
	ThumbsPath      string                         `json:"thumbs_path"`   // path to thumbs on thumb server
	ThumbRetina     bool                           `json:"thumb_retina"`  // there is @2x version of thumb
	ThumbFormat     string                         `json:"thumb_format"`
	ThumbType       string                         `json:"thumb_type"`           // image type ( jpg, webp, png )
	BestThumb       *int16                         `json:"best_thumb,omitempty"` // best thumb indexed from 0
	Type            string                         `json:"type"`
	Priority        int16                          `json:"priority,omitempty"`
	User            ContentResultUser              `json:"user"`
	Categories      TaxonomyResults                `json:"categories,omitempty"`
	Models          TaxonomyResults                `json:"models,omitempty"`
	RotationStatus  *CtrsStatus                    `json:"rotation_status,omitempty"`
	Ctr             *float32                       `json:"ctr,omitempty"`
	Views           int32                          `json:"views"`
	selectedThumb   *int
}

type ContentResults struct {
	Total   int64            `json:"total"`
	From    int              `json:"from"`
	To      int              `json:"to"`
	Page    int              `json:"page"`
	Pages   int              `json:"pages"`
	Items   []*ContentResult `json:"items"`
	Title   string           `json:"title,omitempty"`
	Related []RelatedItem    `json:"related,omitempty"`
}

func (cd ContentDuration) Format() string {
	var d = time.Duration(cd) * time.Second
	d = d.Round(time.Second)
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d", m, s)
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
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + c.ThumbFormat + ".%d." + c.ThumbType
}

func (c *ContentItemResult) Thumb() string {
	return fmt.Sprintf(c.ThumbTemplate(), c.SelectedThumb())
}

func (c *ContentItemResult) HiresThumb() string {
	if c.ThumbRetina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(), "."+c.ThumbType)+"@2x."+c.ThumbType, c.SelectedThumb())
	} else {
		return c.Thumb()
	}
}

func (c ContentItemResult) GalleryInfo(formats ...string) ContentGalleryInfo {
	if c.GalleryItems == nil {
		return ContentGalleryInfo{}
	}
	var galleryInfo ContentGalleryInfo
	var ok bool
	if len(formats) > 0 {
		for _, f := range formats {
			if galleryInfo, ok = (*c.GalleryItems)[f]; ok {
				galleryInfo.Name = f
				break
			}
		}
	}
	if !ok {
		for name, info := range *c.GalleryItems {
			galleryInfo = info
			galleryInfo.Name = name
			break
		}
	}
	return galleryInfo
}
func (c ContentItemResult) GalleryFormats() []string {
	if c.GalleryItems == nil {
		return []string{}
	}
	return lo.Keys(*c.GalleryItems)
}

type GalleryImageInfo struct {
	ImageUrl    string `json:"image_url"`
	PreviewUrl  string `json:"preview_url"`
	PreviewSize Size   `json:"preview_size"`
	ImageSize   Size   `json:"image_size"`
}

func (c ContentItemResult) GalleryImages(formats ...string) (images []GalleryImageInfo) {
	info := c.GalleryInfo(formats...)
	images = make([]GalleryImageInfo, 0, len(info.PreviewItems))
	for k, preview := range info.PreviewItems {
		image := GalleryImageInfo{
			ImageUrl:    fmt.Sprintf("%s%s/image-%s.%d.%s", c.GalleryServer, c.GalleryPath, info.Name, k, info.Type),
			PreviewUrl:  fmt.Sprintf("%s%s/preview-%s.%d.%s", c.GalleryServer, c.GalleryPath, info.Name, k, info.Type),
			PreviewSize: preview,
			ImageSize:   info.Items[k],
		}
		images = append(images, image)
	}
	return
}

func (c ContentItemResult) VideoInfo(formats ...string) ContentVideoInfo {
	if c.VideoSizes == nil {
		return ContentVideoInfo{}
	}
	var videoInfo ContentVideoInfo
	var ok bool
	if len(formats) > 0 {
		for _, f := range formats {
			if videoInfo, ok = (*c.VideoSizes)[f]; ok {
				videoInfo.Name = f
				break
			}
		}
	}
	if !ok {
		for name, info := range *c.VideoSizes {
			videoInfo = info
			videoInfo.Name = name
			break
		}
	}
	return videoInfo
}
func (c ContentItemResult) VideoFormats() []string {
	if c.VideoSizes == nil {
		return []string{}
	}
	return lo.Keys(*c.VideoSizes)
}
func (c ContentItemResult) VideoUrl(formats ...string) string {
	info := c.VideoInfo(formats...)
	return c.VideoServer + c.VideoPath + "/video-" + info.Name + "." + info.Type
}
func (c ContentItemResult) VideoPoster(formats ...string) string {
	info := c.VideoInfo(formats...)
	if info.PosterType == "" {
		return ""
	}
	return c.VideoServer + c.VideoPath + "/poster-" + info.Name + "." + info.PosterType
}
func (c ContentItemResult) VideoTimeline(formats ...string) string {
	info := c.VideoInfo(formats...)
	if info.TimelineType == "" {
		return ""
	}
	return c.VideoServer + c.VideoPath + "/timeline-" + info.Name + ".vtt"
}

func (c ContentItemResult) VideoSize(formats ...string) Size {
	info := c.VideoInfo(formats...)
	return info.Size
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

func (c ContentItemResult) MainCategorySlug(defaultName ...string) string {
	def := "any"
	if len(defaultName) > 0 {
		def = defaultName[0]
	}
	if len(c.Categories) == 0 {
		return def
	}
	return c.Categories[0].Slug
}

func (c ContentResult) ThumbTemplate() string {
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + c.ThumbFormat + ".%d." + c.ThumbType
}

func (c *ContentResult) Thumb() string {
	return fmt.Sprintf(c.ThumbTemplate(), c.SelectedThumb())
}

func (c *ContentResult) HiresThumb() string {
	if c.ThumbRetina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(), "."+c.ThumbType)+"@2x."+c.ThumbType, c.SelectedThumb())
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

func (c ContentResult) MainCategorySlug(defaultName ...string) string {
	def := "any"
	if len(defaultName) > 0 {
		def = defaultName[0]
	}
	if len(c.Categories) == 0 {
		return def
	}
	return c.Categories[0].Slug
}