package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sersh.com/totaltube/frontend/types"
)

func ToJSON(doc interface{}) []byte {
	bt, err := json.Marshal(doc)
	if err != nil {
		return []byte{}
	}
	return bt
}

func DumpJSON(doc interface{}) {
	bt, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(bt))
}

func FromJSON(data []byte, dest interface{}) {
	_ = json.Unmarshal(data, &dest)
}


func Time8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

func Duration8601(d interface{}) string {
	var dd time.Duration
	switch t := d.(type) {
	case types.ContentDuration:
		dd = time.Duration(t)*time.Second
	case int32:
		dd = time.Duration(t)*time.Second
	case int:
		dd = time.Duration(t)*time.Second
	case int64:
		dd = time.Duration(t)*time.Second
	case float64:
		dd = time.Duration(t)*time.Second
	case time.Duration:
		dd = t
	default:
		log.Printf("wrong duration type: %T, value: %v\n", d, d)
	}
	return "PT" + strings.ToUpper(dd.Truncate(time.Millisecond).String())
}