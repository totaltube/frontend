package site

import (
	"github.com/flosch/pongo2/v4"
	"github.com/pkg/errors"
	"log"
	"math"
)

func InitFilters() {
	err := pongo2.RegisterFilter("splitUp", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		if !in.CanSlice() {
			return nil, &pongo2.Error{
				Sender:    "filter:splitUp",
				OrigError: errors.New("splitUp filter must be applied to array/slice"),
			}
		}
		if !param.IsInteger() {
			return nil, &pongo2.Error{
				Sender:    "filter:splitUp",
				OrigError: errors.New("splitUp filter require integer parameter"),
			}
		}
		cols := param.Integer()
		if cols <= 0 {
			return nil, &pongo2.Error{
				Sender:    "filter:splitUp",
				OrigError: errors.New("number of columns for splitUp must be > 0"),
			}
		}
		totalSize := in.Len()
		colSize := int(math.Floor(float64(totalSize) / float64(cols)))
		var inSlice = make([]interface{}, 0, totalSize)
		in.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
			inSlice = append(inSlice, value.Interface())
			return true
		}, nil)
		out = pongo2.AsValue(ChunkSlice(inSlice, colSize))
		return
	})
	if err != nil {
		log.Fatalln(err)
	}
}
