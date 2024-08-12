package db

import (
	"encoding/json"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.etcd.io/bbolt"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

func GetTranslationBolt(from, to, text string) (translation string) {
	keyStr := from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	err := boltdb.Update(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("translations")).Get(key)
		if v == nil {
			return errors.New("not found")
		}
		d, err := Decompress(v)
		if err != nil {
			return err
		}
		translation = string(d)
		if v := tx.Bucket([]byte("translations_access")).Get(key); v == nil {
			// update access time
			_ = tx.Bucket([]byte("translations_access")).Put(key, []byte(""))
			_ = tx.Bucket([]byte("translations_access_expire")).Put(key, []byte(time.Now().Add(updateAccessedTranslationsInterval).Format(time.RFC3339)))
			// update expire time
			_ = tx.Bucket([]byte("translations_expire")).Put(key, []byte(time.Now().Add(expireAccessedTranslationsInterval).Format(time.RFC3339)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return
}

func DeleteTranslationBolt(from, to, text string) {
	keyStr := from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	_ = boltdb.Update(func(tx *bbolt.Tx) error {
		_ = tx.Bucket([]byte("translations")).Delete(key)
		_ = tx.Bucket([]byte("translations_expire")).Delete(key)
		_ = tx.Bucket([]byte("translations_tried")).Delete(key)
		_ = tx.Bucket([]byte("translations_tried_expire")).Delete(key)
		_ = tx.Bucket([]byte("translations_access")).Delete(key)
		_ = tx.Bucket([]byte("translations_access_expire")).Delete(key)
		return nil
	})
}

func SaveTranslationBolt(from, to, text, translation string) {
	key := []byte(from + "_" + to + "_" + helpers.Md5Hash(text))
	_ = boltdb.Update(func(tx *bbolt.Tx) error {
		_ = tx.Bucket([]byte("translations")).Put(key, Compress([]byte(translation)))
		_ = tx.Bucket([]byte("translations_access")).Put(key, []byte(""))
		_ = tx.Bucket([]byte("translations_access_expire")).Put(key, []byte(time.Now().Add(updateAccessedTranslationsInterval).Format(time.RFC3339)))
		_ = tx.Bucket([]byte("translations_expire")).Put(key, []byte(time.Now().Add(expireAccessedTranslationsInterval).Format(time.RFC3339)))
		return nil
	})
}

func SaveDeferredTranslationBolt(from, to, text, Type string) {
	var raceError = errors.New("key already exists")
	keyStr := time.Now().Format(time.RFC3339) + "|" + helpers.RandStr(5)
	key := []byte(keyStr)
	triedKey := []byte(from + "_" + to + "_" + helpers.Md5Hash(text))
	err := boltdb.Update(func(tx *bbolt.Tx) error {
		if v := tx.Bucket([]byte("translations_tried")).Get(triedKey); v != nil {
			// If we already tried to translate this, then not saving anything, waiting when last attempt ttl will expire
			return nil
		}
		if v := tx.Bucket([]byte("translations_deferred")).Get(key); v != nil {
			// race condition - the key is already here, but it shouldn`t
			// trying again then
			return raceError
		}
		_ = tx.Bucket([]byte("translations_tried")).Put(triedKey, []byte(keyStr))
		_ = tx.Bucket([]byte("translations_tried_expire")).Put(triedKey, []byte(time.Now().Add(time.Minute*60).Format(time.RFC3339)))
		_ = tx.Bucket([]byte("translations_deferred")).Put(key, Compress(helpers.ToJSON(translationDoc{
			From: from,
			To:   to,
			Text: text,
			Type: Type,
		})))
		_ = tx.Bucket([]byte("translations_deferred_expire")).Put(key, []byte(time.Now().Add(time.Minute*60).Format(time.RFC3339)))
		return nil
	})
	if errors.Is(err, raceError) {
		time.Sleep(time.Nanosecond)
		SaveDeferredTranslationBolt(from, to, text, Type)
		return
	}
	if err != nil {
		log.Println(err)
	}
}

func doTranslationsBolt() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover in doTranslations:", r)
		}
	}()
	var toTranslate = make([]toTranslateT, 0, 100)
	err := boltdb.View(func(tx *bbolt.Tx) error {
		cursor := tx.Bucket([]byte("translations_deferred")).Cursor()
		from := time.Now().Format(time.RFC3339) + "|" + "aaaaa"
		for k, v := cursor.Seek([]byte(from)); k != nil; k, v = cursor.Next() {
			timestamp := string(k[:len(k)-6])
			now := time.Now()
			if t, err := time.Parse(time.RFC3339, timestamp); err != nil {
				log.Println("wrong timestamp: ", err)
				return err
			} else if t.After(now) {
				break // This will be translated in future
			}
			// translating
			var doc translationDoc
			d, err := Decompress(v)
			if err != nil {
				log.Println(err)
				return err
			}
			err = json.Unmarshal(d, &doc)
			if err != nil {
				log.Println(err)
				return err
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
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	for _, t := range toTranslate {
		translation, err := api.Translate("", t.translate)
		if err != nil {
			log.Printf("Error translating '%s' from %s to %s: %s", t.translate.Text, t.translate.From, t.translate.To, err.Error())
		} else {
			SaveTranslation(t.translate.From, t.translate.To, t.translate.Text, translation)
		}
		_ = boltdb.Update(func(tx *bbolt.Tx) error {
			_ = tx.Bucket([]byte("translations_deferred")).Delete(t.key)
			_ = tx.Bucket([]byte("translations_deferred_expire")).Delete(t.key)
			return nil
		})
	}
}
