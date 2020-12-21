package api

import (
	"github.com/segmentio/encoding/json"
	"log"
	"sersh.com/totaltube/frontend/types"
)

type translateResponse struct {
	Translation string
	Cached      bool
}

func Translate(params types.TranslateParams) (translation string, err error) {
	var response json.RawMessage
	response, err = apiRequest(methodPost, uriTranslate, params)
	if err != nil {
		return
	}
	var tr translateResponse
	err = json.Unmarshal(response, &tr)
	if err != nil {
		log.Println(err, string(response))
		return
	}
	translation = tr.Translation
	return
}
