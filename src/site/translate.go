package site

import "sersh.com/totaltube/frontend/db"

func defferedTranslate(from string, to string, text string) (translation string) {
	translation = db.GetTranslation(from, to, text)
	if translation != "" {
		return
	}
	db.SaveDeferredTranslation(from, to, text) // переведем попозже, как будет время
	return
}
