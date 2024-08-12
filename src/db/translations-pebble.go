package db

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/samber/lo"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

func GetTranslationPebble(from, to, text string) (translation string) {
	keyStr := translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	keyAccess := []byte(translationAccessedPrefix + keyStr)
	batch := pdb.NewIndexedBatch()
	defer batch.Close()
	val, closer, err := batch.Get(key)
	if err != nil {
		return
	}
	translation = string(val)
	closer.Close()
	if _, closer, err = batch.Get(keyAccess); err != nil {
		// update access time
		err = batch.Set(keyAccess, []byte(""), pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
		err = batch.Set([]byte("ie_"+time.Now().Add(updateAccessedTranslationsInterval).Format(time.RFC3339)+"_"+helpers.RandStr(10)), keyAccess, pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
		err = batch.Set([]byte("ie_"+time.Now().Add(expireAccessedTranslationsInterval).Format(time.RFC3339)+"_"+helpers.RandStr(10)), key, pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		closer.Close()
	}
	return
}

func DeleteTranslationPebble(from, to, text string) {
	keyStr := translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	triedKey := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	keyAccess := []byte(translationAccessedPrefix + keyStr)
	_ = pdb.Delete(key, pebble.NoSync)
	_ = pdb.Delete(triedKey, pebble.NoSync)
	_ = pdb.Delete(keyAccess, pebble.NoSync)
}

func SaveTranslationPebble(from, to, text, translation string) {
	key := []byte(translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	err := pdb.Set(key, []byte(translation), pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
	err = pdb.Set([]byte("ie_"+time.Now().Add(expireAccessedTranslationsInterval).Format(time.RFC3339)+"_"+helpers.RandStr(10)), key, pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
}

func SaveDeferredTranslationPebble(from, to, text, Type string) {
	now := time.Now().Format(time.RFC3339Nano)
	key := []byte(translationsDeferredPrefix + now)
	triedKey := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	if _, closer, err := pdb.Get(triedKey); err == nil {
		closer.Close()
		return
	}
	if _, closer, err := pdb.Get(key); err == nil {
		closer.Close()
		// race condition - the key is already here, but it shouldn`t
		// trying again then
		time.Sleep(time.Nanosecond)
		SaveDeferredTranslation(from, to, text, Type)
		return
	}
	batch := pdb.NewBatch()
	defer batch.Close()
	err := batch.Set(triedKey, []byte(""), pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
	err = batch.Set([]byte("ie_"+time.Now().Add(time.Minute*60).Format(time.RFC3339)+"_"+helpers.RandStr(10)), triedKey, pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
	err = batch.Set(key, helpers.ToJSON(translationDoc{
		From: from,
		To:   to,
		Text: text,
		Type: Type,
	}), pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
	err = batch.Set([]byte("ie_"+time.Now().Add(time.Minute*60).Format(time.RFC3339)+"_"+helpers.RandStr(10)), key, pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
	err = batch.Commit(pebble.NoSync)
	if err != nil {
		log.Println(err)
	}
}

func doTranslationsPebble() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover in doTranslations:", r)
		}
	}()
	it, err := pdb.NewIter(&pebble.IterOptions{LowerBound: []byte(translationsDeferredPrefix)})
	if err != nil {
		log.Println(err)
		return
	}
	defer it.Close()
	var toTranslate []toTranslateT
	for valid := it.SeekGE([]byte(translationsDeferredPrefix)); valid; valid = it.Next() {
		k := bytes.Clone(it.Key())
		if !bytes.HasPrefix(k, []byte(translationsDeferredPrefix)) {
			break
		}
		timestamp := bytes.TrimPrefix(k, []byte(translationsDeferredPrefix))
		now := time.Now()
		if t, err := time.Parse(time.RFC3339Nano, string(timestamp)); err != nil {
			log.Println("wrong timestamp: ", err)
			return
		} else if t.After(now) {
			break // This will be translated in future
		}
		// translating
		var doc translationDoc
		v := it.Value()
		err = json.Unmarshal(v, &doc)
		if err != nil {
			log.Println(err)
			return
		}
		Type := doc.Type
		if !lo.Contains(types.TranslationTypes, Type) {
			Type = "page-text"
		}
		toTranslate = append(toTranslate, toTranslateT{
			key: k,
			translate: types.TranslateParams{
				From: doc.From,
				To:   doc.To,
				Text: doc.Text,
				Type: Type,
			},
		})
	}
	for _, t := range toTranslate {
		translation, err := api.Translate("", t.translate)
		if err != nil {
			log.Printf("Error translating '%s' from %s to %s: %s", t.translate.Text, t.translate.From, t.translate.To, err.Error())
		} else {
			SaveTranslationPebble(t.translate.From, t.translate.To, t.translate.Text, translation)
		}
		_ = pdb.Delete(t.key, pebble.NoSync)
	}
}
