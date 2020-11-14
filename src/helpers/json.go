package helpers

import (
	"fmt"
	"github.com/segmentio/encoding/json"
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
