package site

import "sersh.com/totaltube/frontend/db"

func deferredTranslate(from string, to string, text string) (translation string) {
	if from == to {
		return text
	}
	translation = db.GetTranslation(from, to, text)
	if translation != "" {
		return text
	}
	db.SaveDeferredTranslation(from, to, text) // переведем попозже, как будет время
	return
}
