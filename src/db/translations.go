package db

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

const (
	translationsPrefix         = "tr_"
	translationsDeferredPrefix = "trd_"
	translationsTriedPrefix    = "trt_"
	translationAccessedPrefix  = "tra_"
)

const (
	updateAccessedTranslationsInterval = time.Hour * 24          // каждые сутки обновляется TTL переводов, которые были использованы
	expireAccessedTranslationsInterval = time.Hour * 24 * 30 * 6 // через 6 месяцев неиспользованные переводы удаляются
)

type translationDoc struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
	Type string `json:"type"`
}

func GetTranslation(from, to, text string) (translation string) {
	if internal.Config.Database.Engine == "pebble" {
		return GetTranslationPebble(from, to, text)
	}
	if internal.Config.Database.Engine == "bolt" {
		return GetTranslationBolt(from, to, text)
	}
	keyStr := translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	keyAccess := []byte(translationAccessedPrefix + keyStr)
	_ = bdb.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		_ = item.Value(func(val []byte) error {
			translation = string(val)
			return nil
		})
		if _, err = txn.Get(keyAccess); err != nil {
			_ = txn.SetEntry(badger.NewEntry(keyAccess, []byte("")).WithTTL(updateAccessedTranslationsInterval))
			_ = txn.SetEntry(badger.NewEntry(key, []byte(translation)).WithTTL(expireAccessedTranslationsInterval))
		}
		return nil
	})
	return
}

func DeleteTranslation(from, to, text string) {
	if internal.Config.Database.Engine == "pebble" {
		DeleteTranslationPebble(from, to, text)
		return
	}
	if internal.Config.Database.Engine == "bolt" {
		DeleteTranslationBolt(from, to, text)
		return
	}
	keyStr := translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	triedKey := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	keyAccess := []byte(translationAccessedPrefix + keyStr)
	_ = bdb.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(key)
		_ = txn.Delete(triedKey)
		_ = txn.Delete(keyAccess)
		return nil
	})
}

func SaveTranslation(from, to, text, translation string) {
	if internal.Config.Database.Engine == "pebble" {
		SaveTranslationPebble(from, to, text, translation)
		return
	}
	if internal.Config.Database.Engine == "bolt" {
		SaveTranslationBolt(from, to, text, translation)
		return
	}
	key := []byte(translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	_ = bdb.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, []byte(translation)).WithTTL(expireAccessedTranslationsInterval))
	})
}

func SaveDeferredTranslation(from, to, text, Type string) {
	if internal.Config.Database.Engine == "pebble" {
		SaveDeferredTranslationPebble(from, to, text, Type)
		return
	}
	if internal.Config.Database.Engine == "bolt" {
		SaveDeferredTranslationBolt(from, to, text, Type)
		return
	}
	var raceError = errors.New("key already exists")
	now := time.Now().Format(time.RFC3339Nano)
	key := []byte(translationsDeferredPrefix + now)
	triedKey := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	err := bdb.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(triedKey); err == nil {
			// If we already tried to translate this, then not saving anything, waiting when last attempt ttl will expire
			return nil
		}
		if _, err := txn.Get(key); err == nil {
			// race condition - the key is already here, but it shouldn`t
			// trying again then
			return raceError
		}
		_ = txn.SetEntry(badger.NewEntry(triedKey, []byte(now)).WithTTL(time.Minute * 60)) // After one hour will try to translate again
		_ = txn.SetEntry(badger.NewEntry(key, helpers.ToJSON(translationDoc{
			From: from,
			To:   to,
			Text: text,
			Type: Type,
		})).WithTTL(time.Minute * 60))
		return nil
	})
	if errors.Is(err, raceError) {
		time.Sleep(time.Nanosecond)
		SaveDeferredTranslation(from, to, text, Type)
		return
	}
	if err != nil {
		log.Println(err)
	}
}

type toTranslateT struct {
	key       []byte
	translate types.TranslateParams
}

func doTranslations() {
	if internal.Config.Database.Engine == "pebble" {
		doTranslationsPebble()
		return
	}
	if internal.Config.Database.Engine == "bolt" {
		doTranslationsBolt()
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover in doTranslations:", r)
		}
	}()
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
				break // This will be translated in future
			} else {
				// translating
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
		}
		for _, t := range toTranslate {
			translation, err := api.Translate("", t.translate)
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
