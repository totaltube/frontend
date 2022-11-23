package types

import (
	"testing"
)

func TestCategoryResult_SelectedThumb(t *testing.T) {
	cat := &CategoryResult{
		ThumbRetina:   true,
		ThumbWidth:    0,
		ThumbHeight:   0,
		ThumbsAmount:  5,
		ThumbsServer:  "http://someserver.com",
		ThumbsPath:    "/some/path",
		ThumbFormat:   "main",
		selectedThumb: nil,
	}
	var v interface{} = cat
	t.Log(cat.SelectedThumb())
	r := v.(Thumber)
	t.Log(r.SelectedThumb())
}
