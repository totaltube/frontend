package db

import (
	"bytes"
	"encoding/json"
	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"log"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
	"time"
)

const (
	translationsPrefix         = "tr_"
	translationsDeferredPrefix = "trd_"
	translationsTriedPrefix    = "trt_"
)

type translationDoc struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

func GetTranslation(from, to, text string) (translation string) {
	key := []byte(translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	_ = bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		_ = item.Value(func(val []byte) error {
			translation = string(val)
			return nil
		})
		return nil
	})
	return
}

func SaveTranslation(from, to, text, translation string) {
	key := []byte(translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	_ = bdb.Update(func(txn *badger.Txn) error {
		return txn.Set(key, []byte(translation))
	})
}

func SaveDeferredTranslation(from, to, text string) {
	var raceError = errors.New("key already exists")
	now := time.Now().Format(time.RFC3339Nano)
	key := []byte(translationsDeferredPrefix + now)
	triedKey := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	err := bdb.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(triedKey); err == nil {
			// Если уже пытались перевести это дело, то ничего не сохраняем, а ждем когда пройдет ttl последней попытки
			return nil
		}
		if _, err := txn.Get(key); err == nil {
			// race condition - ключ уже есть, хотя он по идее не должен быть, ибо нанотайм
			// тогда попробуем еще раз
			return raceError
		}
		_ = txn.SetEntry(badger.NewEntry(triedKey, []byte(now)).WithTTL(time.Minute * 60)) // Минимум раз в час пробуем еще раз перевести
		_ = txn.SetEntry(badger.NewEntry(key, helpers.ToJSON(translationDoc{
			From: from,
			To:   to,
			Text: text,
		})).WithTTL(time.Minute * 60))
		return nil
	})
	if err == raceError {
		time.Sleep(time.Nanosecond)
		SaveDeferredTranslation(from, to, text)
		return
	}
	if err != nil {
		log.Println(err)
	}
}
type toTranslateT struct {
	key []byte
	translate types.TranslateParams
}

func doTranslations() {
	_ = bdb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(translationsDeferredPrefix)
		var toTranslate = make([]toTranslateT, 0, 100)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			timestamp := bytes.TrimPrefix(k, prefix)
			now := time.Now()
			if t, err := time.Parse(time.RFC3339Nano, string(timestamp)); err != nil {
				log.Println("wrong timestamp: ", err)
				return err
			} else if t.After(now) {
				log.Println(t, now)
				break // Это мы переведем в будущем
			} else {
				// переводим
				var doc translationDoc
				err = item.Value(func(val []byte) error {
					err = json.Unmarshal(val, &doc)
					if err != nil {
						log.Println(err)
						return err
					}
					return nil
				})
				if err != nil {
					log.Println(err)
					return err
				}
				toTranslate = append(toTranslate, toTranslateT{
					key:       k,
					translate: types.TranslateParams{
						From: doc.From,
						To: doc.To,
						Text: doc.Text,
					},
				})
			}
		}
		for _, t := range toTranslate {
			translation, err := api.Translate(t.translate)
			if err != nil {
				log.Printf("Error translating '%s' from %s to %s: %s", t.translate.Text, t.translate.From, t.translate.To, err.Error())
			} else {
				SaveTranslation(t.translate.From, t.translate.To, t.translate.Text, translation)
			}
			_ = bdb.Update(func(txn *badger.Txn) error {
				return txn.Delete(t.key)
			})
		}
		return nil
	})
}
