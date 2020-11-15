package types

type Thumber interface {
	ThumbTemplate() string
	Thumb() string
	HiresThumb() string
	SelectedThumb() int
}
