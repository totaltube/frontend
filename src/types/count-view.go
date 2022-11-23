package types

type CountViewParams struct {
	Type    string
	Id      int64
	Ip      string
	Slug    string
	ThumbId int16 `json:"thumb_id"`
}
