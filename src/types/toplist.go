package types

type ToplistContentData struct {
	// params needed to identify the content item and build the link to it
	ContentId int64  `json:"content_id"`    // optional
	Url       string `json:"url,omitempty"` // optional
}
type ToplistItem struct {
	// params needed to display the content item as toplist item
	Thumb       string `json:"thumb"`
	HiresThumb  string `json:"hires_thumb,omitempty"` // optional
	Title       string `json:"title,omitempty"`       // optional
	Description string `json:"description,omitempty"` // optional

	ContentData ToplistContentData `json:"content_data"`
}
type ToplistResults struct {
	Success bool          `json:"success"`
	Items   []ToplistItem `json:"items"`
}
