package types

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
