package site

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/pkg/errors"
	"log"
	"math"
	"sersh.com/totaltube/frontend/types"
)

func InitPongo2() {
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

	err = pongo2.RegisterFilter("translate", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		if !in.IsString() || in.String() == "" {
			return nil, &pongo2.Error{
				Sender:    "filter:translate",
				OrigError: errors.New("translate filter must be applied to string"),
			}
		}
		if !param.IsString() || param.String() == "" {
			return nil, &pongo2.Error{
				Sender:    "filter:translate",
				OrigError: errors.New("translate filter needs one param - language code to translate to"),
			}
		}
		out = pongo2.AsValue(deferredTranslate("en", param.String(), in.String()))
		return
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = pongo2.RegisterFilter("thumb", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		if it, ok := in.Interface().(types.Thumber); ok {
			out = pongo2.AsValue(it.Thumb())
			return
		} else {
			return nil, &pongo2.Error{
				Sender:    "filter:thumb",
				OrigError: errors.New(fmt.Sprintf("wrong item type: %T", in.Interface())),
			}
		}
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = pongo2.RegisterFilter("thumb_template", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		it := in.Interface().(types.Thumber)
		out = pongo2.AsValue(it.ThumbTemplate())
		return
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = pongo2.RegisterTag("dynamic", pongo2Dynamic)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("fetch", pongo2Fetch)
	if err != nil {
		log.Fatalln(err)
	}
}
