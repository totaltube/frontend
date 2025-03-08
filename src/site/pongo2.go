package site

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/types"
)

func InitPongo2() {
	err := pongo2.RegisterFilter("dump", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		bt, err1 := json.MarshalIndent(in.Interface(), "", "  ")
		if err1 != nil {
			return nil, &pongo2.Error{
				Sender:    "filter:dump",
				OrigError: err1,
			}
		}
		out = pongo2.AsValue(string(bt))
		return
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterFilter("splitUp", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		defer func() {
			if r := recover(); r != nil {
				err = &pongo2.Error{
					Sender:    "filter:splitUp",
					OrigError: errors.New(fmt.Sprintf("%s", r)),
				}
			}
		}()
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
		colSize := int(math.Ceil(float64(totalSize) / float64(cols)))
		var inSlice = make([]interface{}, 0, totalSize)
		in.Iterate(func(idx, count int, value, _ *pongo2.Value) bool {
			inSlice = append(inSlice, value.Interface())
			return true
		}, func() {
			log.Println("empty")
		})
		out = pongo2.AsValue(ChunkSlice(inSlice, colSize))
		return
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterFilter("from", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		defer func() {
			if r := recover(); r != nil {
				err = &pongo2.Error{
					Sender:    "filter:from",
					OrigError: errors.New(fmt.Sprintf("%s", r)),
				}
			}
		}()
		if !in.CanSlice() {
			return nil, &pongo2.Error{
				Sender:    "filter:from",
				OrigError: errors.New("from filter must be applied to array/slice"),
			}
		}
		if !param.IsInteger() {
			return nil, &pongo2.Error{
				Sender:    "filter:from",
				OrigError: errors.New("form filter require integer parameter"),
			}
		}
		from := param.Integer()
		totalSize := in.Len()
		if from >= totalSize-1 {
			out = pongo2.AsValue([]interface{}{})
			return
		}
		var outSlice = make([]interface{}, 0, totalSize-from)
		in.Iterate(func(idx, count int, value, _ *pongo2.Value) bool {
			if idx >= from {
				outSlice = append(outSlice, value.Interface())
			}
			return true
		}, func() {
		})
		out = pongo2.AsValue(outSlice)
		return
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = pongo2.RegisterFilter("max", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		defer func() {
			if r := recover(); r != nil {
				err = &pongo2.Error{
					Sender:    "filter:max",
					OrigError: errors.New(fmt.Sprintf("%s", r)),
				}
			}
		}()
		if !in.CanSlice() {
			return nil, &pongo2.Error{
				Sender:    "filter:max",
				OrigError: errors.New("max filter must be applied to array/slice"),
			}
		}
		if !param.IsInteger() {
			return nil, &pongo2.Error{
				Sender:    "filter:max",
				OrigError: errors.New("max filter require integer parameter"),
			}
		}
		max := param.Integer()
		totalSize := in.Len()
		var outSlice = make([]interface{}, 0, int(math.Min(float64(totalSize), float64(max))))

		in.Iterate(func(idx, count int, value, _ *pongo2.Value) bool {
			if idx < max {
				outSlice = append(outSlice, value.Interface())
			}
			return true
		}, func() {
		})
		out = pongo2.AsValue(outSlice)
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
		out = pongo2.AsValue(deferredTranslate("en", param.String(), in.String(), "page-text", false, nil))
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

	err = pongo2.RegisterFilter("boolean", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		var res = false
		if in.IsBool() {
			res = in.Bool()
		} else if in.IsString() {
			res, _ = strconv.ParseBool(in.String())
		} else if in.IsNumber() {
			res = in.Integer() != 0
		} else if in.CanSlice() {
			res = in.Len() > 0
		}
		if res {
			out = pongo2.AsValue("true")
		} else {
			out = pongo2.AsValue("false")
		}
		return
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterFilter("string", func(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
		out = pongo2.AsValue(fmt.Sprintf("%v", in.Interface()))
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
	err = pongo2.RegisterTag("link", pongo2Link)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("canonical", pongo2Canonical)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("alternates", pongo2Alternates)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("alternate", pongo2Alternate)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("page_link", pongo2PageLink)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("cdata", pongo2CData)
	err = pongo2.RegisterTag("prevnext", pongo2Prevnext)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("sort", pongo2Sort)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.ReplaceTag("set", pongo2Set)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("repeat", pongo2Repeat)
	if err != nil {
		log.Fatalln(err)
	}
	err = pongo2.RegisterTag("dilute", pongo2Dilute)
	if err != nil {
		log.Fatalln(err)
	}
}
