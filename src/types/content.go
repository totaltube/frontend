package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
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

type CustomData map[string]interface{}
type CustomTranslations map[string]string

type ThumbFormat struct {
	Name                string   `json:"name"`
	Width               int64    `json:"width"`
	Height              int64    `json:"height"`
	Amount              int64    `json:"amount"`
	Type                string   `json:"type"`
	Retina              bool     `json:"retina"`
	VideoPreviewFormats []string `json:"video_preview_formats"`
	VideoPreviewSize    Size     `json:"video_preview_size"`
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

type ChannelShortResult struct {
	Id     int32  `json:"id"`
	Slug   string `json:"slug"`
	Title  string `json:"title"`
	Url    string `json:"url"`
	Banner string `json:"banner"`
}
type RatingsResults map[string]RatingsResult
type RatingsResult struct {
	Likes    int64 `json:"likes"`
	Dislikes int64 `json:"dislikes"`
}

type TaxonomyResults []TaxonomyResult

type ContentResultUser struct {
	Id    int32  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}
type ContentItemResult struct {
	Id                 int64                          `json:"id"`
	Slug               string                         `json:"slug"`
	Title              string                         `json:"title"`
	TitleTranslated    bool                           `json:"title_translated,omitempty"`
	OriginalTitle      string                         `json:"original_title"`
	Description        *string                        `json:"description,omitempty"`
	Channel            *ChannelShortResult            `json:"channel,omitempty"`
	Content            *string                        `json:"content,omitempty"`
	Link               *string                        `json:"link,omitempty"`
	CreatedAt          time.Time                      `json:"created_at"`
	Dated              time.Time                      `json:"dated"`
	Duration           ContentDuration                `json:"duration"`
	Tags               []string                       `json:"tags"`
	Keywords           []string                       `json:"keywords,omitempty"`
	VideoServer        string                         `json:"video_server,omitempty"`
	GalleryServer      string                         `json:"gallery_server,omitempty"`
	VideoPath          string                         `json:"video_path,omitempty"`
	GalleryPath        string                         `json:"gallery_path,omitempty"`
	VideoPreviewServer string                         `json:"video_preview_server,omitempty"`
	VideoPreviewPath   string                         `json:"video_preview_path,omitempty"`
	PosterServer       string                         `json:"poster_server,omitempty"`
	PosterPath         string                         `json:"poster_path,omitempty"`
	GalleryItems       *map[string]ContentGalleryInfo `json:"gallery_items,omitempty"`
	VideoSizes         *map[string]ContentVideoInfo   `json:"video_sizes,omitempty"`
	ThumbFormats       []ThumbFormat                  `json:"thumb_formats"`
	ThumbsServer       string                         `json:"thumbs_server"`        // thumb server url
	ThumbsPath         string                         `json:"thumbs_path"`          // path to thumbs on thumb server
	ThumbRetina        bool                           `json:"thumb_retina"`         // deprecated
	ThumbWidth         int32                          `json:"thumb_width"`          // deprecated
	ThumbsWidth        int32                          `json:"thumbs_width"`         // deprecated
	ThumbHeight        int32                          `json:"thumb_height"`         // deprecated
	ThumbsHeight       int32                          `json:"thumbs_height"`        // deprecated
	ThumbsAmount       int32                          `json:"thumbs_amount"`        // deprecated
	ThumbFormat        string                         `json:"thumb_format"`         // deprecated
	ThumbType          string                         `json:"thumb_type"`           // deprecated
	BestThumb          *int16                         `json:"best_thumb,omitempty"` // best thumb indexed from 0
	Type               string                         `json:"type"`
	Priority           int16                          `json:"priority,omitempty"`
	User               ContentResultUser              `json:"user"`
	Categories         TaxonomyResults                `json:"categories,omitempty"`
	Models             TaxonomyResults                `json:"models,omitempty"`
	Views              int32                          `json:"views"`
	Related            []*ContentResult               `json:"related,omitempty"` // similar content
	SourceSiteId       string                         `json:"source_site_id"`
	SourceSiteUniqueId string                         `json:"source_site_unique_id"`
	CustomData         CustomData                     `json:"custom_data"`
	CustomTranslations CustomTranslations             `json:"custom_translations"`
	Ratings            RatingsResults                 `json:"ratings"`
	selectedThumb      *int
}

type ContentResult struct {
	Id                 int64                          `json:"id"`
	Slug               string                         `json:"slug"`
	Title              string                         `json:"title"`
	TitleTranslated    bool                           `json:"title_translated,omitempty"`
	OriginalTitle      string                         `json:"original_title"`
	Description        *string                        `json:"description,omitempty"`
	Channel            *ChannelShortResult            `json:"channel,omitempty"`
	Content            *string                        `json:"content,omitempty"`
	Link               *string                        `json:"link,omitempty"`
	CreatedAt          time.Time                      `json:"created_at"`
	Dated              time.Time                      `json:"dated"`
	Duration           ContentDuration                `json:"duration"`
	Tags               []string                       `json:"tags"`
	Keywords           []string                       `json:"keywords,omitempty"`
	GalleryServer      string                         `json:"gallery_server"`
	VideoServer        string                         `json:"video_server"`
	GalleryPath        string                         `json:"gallery_path"`
	VideoPath          string                         `json:"video_path"`
	GalleryItems       *map[string]ContentGalleryInfo `json:"gallery_items,omitempty"`
	VideoSizes         *map[string]ContentVideoInfo   `json:"video_sizes,omitempty"`
	ThumbsServer       string                         `json:"thumbs_server"` // thumb server url
	ThumbsPath         string                         `json:"thumbs_path"`   // path to thumbs on thumb server
	VideoPreviewServer string                         `json:"video_preview_server,omitempty"`
	VideoPreviewPath   string                         `json:"video_preview_path,omitempty"`
	PosterServer       string                         `json:"poster_server,omitempty"`
	PosterPath         string                         `json:"poster_path,omitempty"`
	ThumbFormats       []ThumbFormat                  `json:"thumb_formats"`
	ThumbRetina        bool                           `json:"thumb_retina"`         // deprecated
	ThumbWidth         int32                          `json:"thumb_width"`          // deprecated
	ThumbsWidth        int32                          `json:"thumbs_width"`         // deprecated
	ThumbHeight        int32                          `json:"thumb_height"`         // deprecated
	ThumbsHeight       int32                          `json:"thumbs_height"`        // deprecated
	ThumbsAmount       int32                          `json:"thumbs_amount"`        // deprecated
	ThumbFormat        string                         `json:"thumb_format"`         // deprecated
	ThumbType          string                         `json:"thumb_type"`           // deprecated
	BestThumb          *int16                         `json:"best_thumb,omitempty"` // best thumb indexed from 0
	Type               string                         `json:"type"`
	Priority           int16                          `json:"priority,omitempty"`
	User               ContentResultUser              `json:"user"`
	Categories         TaxonomyResults                `json:"categories,omitempty"`
	Models             TaxonomyResults                `json:"models,omitempty"`
	RotationStatus     *CtrsStatus                    `json:"rotation_status,omitempty"`
	Ctr                *float32                       `json:"ctr,omitempty"`
	Views              int32                          `json:"views"`
	SourceSiteId       string                         `json:"source_site_id"`
	SourceSiteUniqueId string                         `json:"source_site_unique_id"`
	CustomData         CustomData                     `json:"custom_data"`
	CustomTranslations CustomTranslations             `json:"custom_translations"`
	Ratings            RatingsResults                 `json:"ratings"`
	selectedThumb      *int
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
	if d < time.Hour {
		m := d / time.Minute
		d -= m * time.Minute
		s := d / time.Second
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
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

func (c ContentItemResult) HasCustomField(name string) bool {
	if c.CustomData == nil {
		return false
	}
	_, ok := c.CustomData[name]
	return ok
}

func (c ContentItemResult) CustomField(name string) interface{} {
	if c.CustomData == nil {
		return nil
	}
	if data, ok := c.CustomData[name]; ok {
		return data
	}
	return nil
}

func (c ContentItemResult) HasCustomTranslation(key string) bool {
	if c.CustomTranslations == nil {
		return false
	}
	_, ok := c.CustomTranslations[key]
	return ok
}

func (c ContentItemResult) CustomTranslation(key string) (translation string) {
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

func (c ContentItemResult) GetThumbFormat(thumbFormatName ...string) (res ThumbFormat) {
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

func (c ContentItemResult) ThumbTemplate(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + format.Name + ".%d." + format.Type
}

func (c *ContentItemResult) Thumb(thumbFormatName ...string) string {
	return fmt.Sprintf(c.ThumbTemplate(thumbFormatName...), c.SelectedThumb(thumbFormatName...))
}

func (c *ContentItemResult) HiresThumb(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	if format.Retina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(thumbFormatName...), "."+format.Type)+"@2x."+format.Type, c.SelectedThumb(thumbFormatName...))
	} else {
		return c.Thumb(thumbFormatName...)
	}
}

func (c *ContentItemResult) GetVideoPreview(thumbFormatName ...string) (res string) {
	if len(c.ThumbFormats) == 0 {
		return
	}
	thumbFormat := ""
	videoPreviewType := ""
	if len(thumbFormatName) > 0 {
		thumbFormat = thumbFormatName[0]
	}
	if len(thumbFormatName) > 1 {
		videoPreviewType = thumbFormatName[1]
	}
	var format ThumbFormat
	for _, f := range c.ThumbFormats {
		if len(f.VideoPreviewFormats) > 0 {
			format = f
			if videoPreviewType != "" && !lo.Contains(f.VideoPreviewFormats, videoPreviewType) {
				continue
			}
		} else {
			continue
		}
		if thumbFormat != "" {
			if f.Name == thumbFormat {
				if videoPreviewType != "" && !lo.Contains(f.VideoPreviewFormats, videoPreviewType) {
					videoPreviewType = f.VideoPreviewFormats[0]
					break
				}
			}
		}
	}
	if format.Name == "" {
		return
	}
	serverPath := c.VideoPreviewServer + c.VideoPreviewPath
	if serverPath == "" {
		serverPath = c.ThumbsServer + c.ThumbsPath
	}
	res = serverPath + "/video-preview-" + format.Name + "." + videoPreviewType
	return
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

func (c ContentItemResult) Mp4VideoFormats() []string {
	if c.VideoSizes == nil {
		return []string{}
	}
	var formats []string
	for name, info := range *c.VideoSizes {
		if info.Type == "mp4" {
			formats = append(formats, name)
		}
	}
	return formats
}
func (c ContentItemResult) VideoUrl(formats ...string) string {
	info := c.VideoInfo(formats...)
	return c.VideoServer + c.VideoPath + "/video-" + info.Name + "." + info.Type
}

func (c ContentItemResult) HlsMasterUrl() string {
	formats := c.Mp4VideoFormats()
	if len(formats) == 0 {
		return ""
	}
	formatNames := strings.Join(formats, ",")
	return c.VideoServer + c.VideoPath + "/video-," + formatNames + ",.mp4.urlset/master.m3u8"
}

func (c ContentItemResult) VideoPoster(formats ...string) string {
	var info ContentVideoInfo
	if c.VideoSizes != nil {
		var lastSize int64
		for name, i := range *c.VideoSizes {
			if len(formats) > 0 {
				if !lo.Contains(formats, i.Name) {
					continue
				}
			}
			if i.PosterType != "" && lastSize < i.Size.Width+i.Size.Height {
				info = i
				info.Name = name
				lastSize = i.Size.Width + i.Size.Height
			}
		}
	}
	if info.PosterType == "" {
		return ""
	}
	serverPath := c.PosterServer + c.PosterPath
	if serverPath == "" {
		serverPath = c.VideoServer + c.VideoPath
	}
	return serverPath + "/poster-" + info.Name + "." + info.PosterType
}

func (c ContentItemResult) VideoTimeline(formats ...string) string {
	var info ContentVideoInfo
	if c.VideoSizes != nil {
		for name, i := range *c.VideoSizes {
			if len(formats) > 0 && !lo.Contains(formats, i.Name) {
				continue
			}
			if i.TimelineType != "" {
				info = i
				info.Name = name
				break
			}
		}
	}
	if info.TimelineType == "" {
		return ""
	}
	return c.VideoServer + c.VideoPath + "/timeline-" + info.Name + ".vtt"
}

func (c ContentItemResult) MaxVideoSize() Size {
	var info ContentVideoInfo
	var lastSize int64
	if c.VideoSizes != nil {
		for name, i := range *c.VideoSizes {
			if i.Size.Width+i.Size.Height > lastSize {
				info = i
				info.Name = name
				lastSize = i.Size.Width + i.Size.Height
			}
		}
	}
	return info.Size
}

func (c ContentItemResult) VideoSize(formats ...string) Size {
	info := c.VideoInfo(formats...)
	return info.Size
}

func (c *ContentItemResult) SelectedThumb(thumbFormatName ...string) (selectedThumb int) {
	maxAmount := int64(-1)
	for _, format := range c.ThumbFormats {
		if format.Amount < maxAmount || maxAmount == -1 {
			maxAmount = format.Amount
		}
	}
	if c.selectedThumb != nil {
		selectedThumb = *c.selectedThumb
		return
	}
	if c.BestThumb != nil {
		idx := int(*c.BestThumb)
		c.selectedThumb = &idx
	} else {
		idx := rand.Intn(int(maxAmount))
		c.selectedThumb = &idx
	}
	selectedThumb = *c.selectedThumb
	return
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

func (c ContentResult) HasCustomField(name string) bool {
	if c.CustomData == nil {
		return false
	}
	_, ok := c.CustomData[name]
	return ok
}

func (c ContentResult) CustomField(name string) interface{} {
	if c.CustomData == nil {
		return nil
	}
	if data, ok := c.CustomData[name]; ok {
		return data
	}
	return nil
}

func (c ContentResult) HasCustomTranslation(key string) bool {
	if c.CustomTranslations == nil {
		return false
	}
	_, ok := c.CustomTranslations[key]
	return ok
}

func (c ContentResult) CustomTranslation(key string) (translation string) {
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

func (c ContentResult) GetThumbFormat(thumbFormatName ...string) (res ThumbFormat) {
	if len(c.ThumbFormats) == 0 {
		return
	}
	res = c.ThumbFormats[0]
	if len(thumbFormatName) > 0 {
		for _, name := range thumbFormatName {
			name := name
			if f, ok := lo.Find(c.ThumbFormats, func(tf ThumbFormat) bool { return tf.Name == name }); ok {
				res = f
				break
			}
		}
	}
	return
}

func (c ContentResult) ThumbTemplate(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	return c.ThumbsServer + c.ThumbsPath + "/thumb-" + format.Name + ".%d." + format.Type
}

func (c *ContentResult) Thumb(thumbFormatName ...string) string {
	return fmt.Sprintf(c.ThumbTemplate(thumbFormatName...), c.SelectedThumb(thumbFormatName...))
}

func (c *ContentResult) HiresThumb(thumbFormatName ...string) string {
	format := c.GetThumbFormat(thumbFormatName...)
	if format.Retina {
		return fmt.Sprintf(strings.TrimSuffix(c.ThumbTemplate(thumbFormatName...), "."+format.Type)+"@2x."+format.Type, c.SelectedThumb(thumbFormatName...))
	} else {
		return c.Thumb(thumbFormatName...)
	}
}
func (c *ContentResult) GetVideoPreview(thumbFormatName ...string) (res string) {
	if len(c.ThumbFormats) == 0 {
		return
	}
	thumbFormat := ""
	videoPreviewType := ""
	if len(thumbFormatName) > 0 {
		thumbFormat = thumbFormatName[0]
	}
	if len(thumbFormatName) > 1 {
		videoPreviewType = thumbFormatName[1]
	}
	var format ThumbFormat
	for _, f := range c.ThumbFormats {
		if len(f.VideoPreviewFormats) > 0 {
			format = f
			if videoPreviewType != "" && !lo.Contains(f.VideoPreviewFormats, videoPreviewType) {
				continue
			}
		} else {
			continue
		}
		if thumbFormat != "" {
			if f.Name == thumbFormat {
				if videoPreviewType != "" && !lo.Contains(f.VideoPreviewFormats, videoPreviewType) {
					videoPreviewType = f.VideoPreviewFormats[0]
					break
				}
			}
		}
	}
	if format.Name == "" {
		return
	}
	if videoPreviewType == "" {
		videoPreviewType = format.VideoPreviewFormats[0]
	}
	serverPath := c.VideoPreviewServer + c.VideoPreviewPath
	if serverPath == "" {
		serverPath = c.ThumbsServer + c.ThumbsPath
	}
	res = serverPath + "/video-preview-" + format.Name + "." + videoPreviewType
	return
}

func (c *ContentResult) SelectedThumb(thumbFormatName ...string) int {
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
