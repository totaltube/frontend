package site

import (
	"fmt"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

func deferredTranslate(from string, to string, text interface{}, Type string, refresh bool, config *types.Config) (translation interface{}) {
	var readyText string
	switch t := text.(type) {
	case *string:
		if t != nil {
			readyText = *t
		}
	case string:
		readyText = t
	case []string:
		result := make([]string, 0, len(t))
		for _, s := range t {
			if s == "" {
				result = append(result, "")
				continue
			}
			if refresh {
				db.DeleteTranslation(from, to, s)
			}
			if config != nil {
				if tr, ok := config.Translations[to]; ok {
					if trr, okk := tr[s]; okk {
						result = append(result, trr)
						continue
					}
				}
			}
			if tr, ok := internal.Config.Translations[to]; ok {
				if trr, okk := tr[s]; okk {
					result = append(result, trr)
					continue
				}
			}
			stringTranslation := db.GetTranslation(from, to, s)
			if stringTranslation == "" {
				db.SaveDeferredTranslation(from, to, s, Type)
				result = append(result, s)
				continue
			}
			result = append(result, stringTranslation)
		}
		translation = result
		return
	default:
		readyText = fmt.Sprintf("%v", t)
	}
	if from == to {
		return readyText
	}
	if readyText == "" || readyText == "<nil>" {
		return ""
	}
	if refresh {
		db.DeleteTranslation(from, to, readyText)
	}
	if config != nil {
		if tr, ok := config.Translations[to]; ok {
			if trr, okk := tr[readyText]; okk {
				translation = trr
				return
			}
		}
	}
	if tr, ok := internal.Config.Translations[to]; ok {
		if trr, okk := tr[readyText]; okk {
			translation = trr
			return
		}
	}
	translation = db.GetTranslation(from, to, readyText)
	if translation == "" {
		db.SaveDeferredTranslation(from, to, readyText, Type) // will translate later
		return readyText
	}
	return
}
