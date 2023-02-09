package site

import (
	"fmt"

	"sersh.com/totaltube/frontend/db"
)

func deferredTranslate(from string, to string, text interface{}, Type string, refresh bool) (translation string) {
	var readyText string
	switch t := text.(type) {
	case *string:
		if t != nil {
			readyText = *t
		}
	case string:
		readyText = t
	default:
		readyText = fmt.Sprintf("%v", t)
	}
	if from == to {
		return readyText
	}
	if readyText == "" || readyText == "<nil>"{
		return ""
	}
	if refresh {
		db.DeleteTranslation(from, to, readyText)
	}
	translation = db.GetTranslation(from, to, readyText)
	if translation == "" {
		db.SaveDeferredTranslation(from, to, readyText, Type) // will translate later
		return readyText
	}
	return
}
