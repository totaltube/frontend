package db

import (
	"encoding/json"
	"log"
	"regexp"
	"sync"
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
	translationsPrefix             = "tr_"
	translationsDeferredPrefix     = "trd_"
	translationsDeferredPagePrefix = "trdp_"
	translationsTriedPrefix        = "trt_"
	translationAccessedPrefix      = "tra_"
)

const (
	updateAccessedTranslationsInterval = time.Hour * 24          // каждые сутки обновляется TTL переводов, которые были использованы
	expireAccessedTranslationsInterval = time.Hour * 24 * 30 * 6 // через 6 месяцев неиспользованные переводы удаляются
)

var timeRegex = regexp.MustCompile(`trdp?__?(.+?)(_|$)`)

type translationDoc struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
	Type string `json:"type"`
}

func GetTranslation(from, to, text string) (translation string) {
	keyStr := translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	keyAccess := []byte(translationAccessedPrefix + keyStr)
	if internal.Config.Database.NoTranslationsAccessUpdate {
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
	keyStr := translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text)
	key := []byte(keyStr)
	keyExists := []byte(translationsDeferredPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	triedKey := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	keyAccess := []byte(translationAccessedPrefix + keyStr)
	_ = bdb.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(key)
		_ = txn.Delete(triedKey)
		_ = txn.Delete(keyAccess)
		_ = txn.Delete(keyExists)
		return nil
	})
}

func SaveTranslation(from, to, text, translation string) {
	key := []byte(translationsPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	_ = bdb.Update(func(txn *badger.Txn) error {
		if internal.Config.Database.NoTranslationsAccessUpdate {
			return txn.SetEntry(badger.NewEntry(key, []byte(translation)))
		}
		return txn.SetEntry(badger.NewEntry(key, []byte(translation)).WithTTL(expireAccessedTranslationsInterval))
	})
}

var ErrExists = errors.New("translation already added")

func SaveDeferredTranslation(from, to, text, Type string) {
	if !lo.Contains(types.TranslationTypes, Type) {
		Type = "page-text"
	}
	now := time.Now().Format(time.RFC3339)
	var keyExists, key, triedKey []byte
	if Type == "page-text" {
		keyExists = []byte(translationsDeferredPagePrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
		key = []byte(translationsDeferredPagePrefix + now + "_" + from + "_" + to + "_" + helpers.Md5Hash(text))
		triedKey = []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	} else {
		keyExists = []byte(translationsDeferredPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
		key = []byte(translationsDeferredPrefix + now + "_" + from + "_" + to + "_" + helpers.Md5Hash(text))
		triedKey = []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	}
	err := bdb.View(func(txn *badger.Txn) error {
		if _, err := txn.Get(keyExists); err == nil {
			// We already have this translation deferred. No need to add more
			return ErrExists
		}
		if _, err := txn.Get(triedKey); err == nil {
			// If we already tried to translate this, then not saving anything, waiting when last attempt ttl will expire
			return ErrExists
		}
		return nil
	})
	if err == ErrExists {
		return
	}
	_ = bdb.Update(func(txn *badger.Txn) error {
		_ = txn.SetEntry(badger.NewEntry(keyExists, []byte("")).WithTTL(time.Minute * 60))
		_ = txn.SetEntry(badger.NewEntry(key, helpers.ToJSON(translationDoc{
			From: from,
			To:   to,
			Text: text,
			Type: Type,
		})).WithTTL(time.Minute * 60))
		return nil
	})
}

type toTranslateT struct {
	key       []byte
	translate types.TranslateParams
}

func TryAgainTranslation(from, to, text string) {
	key := []byte(translationsTriedPrefix + from + "_" + to + "_" + helpers.Md5Hash(text))
	_ = bdb.Update(func(txn *badger.Txn) error {
		item := badger.NewEntry(key, []byte(time.Now().Format(time.RFC3339Nano))).WithTTL(time.Minute * 60)
		return txn.SetEntry(item)
	})
}

var ErrBreak = errors.New("break")

func doTranslations() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recover in doTranslations:", r)
		}
	}()
	var toTranslate = make([]toTranslateT, 0, 1000)
	_ = bdb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix1 := []byte(translationsDeferredPagePrefix)
		prefix2 := []byte(translationsDeferredPrefix)
		process := func(it *badger.Iterator) (err error) {
			item := it.Item()
			k := item.Key()
			now := time.Now()
			matches := timeRegex.FindSubmatch(k)
			if matches == nil {
				log.Println("wrong key: ", string(k))
				return nil
			}
			var t time.Time
			t, err = time.Parse(time.RFC3339, string(matches[1]))
			if err != nil {
				t, err = time.Parse(time.RFC3339Nano, string(matches[1]))
			}
			if err != nil {
				return nil
			}
			if t.After(now) {
				log.Println(t, now)
				return ErrBreak // This will be translated in future
			} else {
				// translating
				var doc translationDoc
				err = item.Value(func(val []byte) (err error) {
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

				if _, err := txn.Get([]byte(translationsTriedPrefix + doc.From + "_" + doc.To + "_" + helpers.Md5Hash(doc.Text))); err == nil {
					// If we already tried to translate this, then not saving anything, waiting when last attempt ttl will expire
					return nil
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
				if len(toTranslate) >= 1000 {
					log.Println("toTranslate >= 1000")
					return ErrBreak
				}
			}
			return
		}
		for it.Seek(prefix1); it.ValidForPrefix(prefix1); it.Next() {
			err := process(it)
			if err == ErrBreak {
				break
			}
		}
		it.Rewind()
		for it.Seek(prefix2); it.ValidForPrefix(prefix2); it.Next() {
			err := process(it)
			if err == ErrBreak {
				break
			}
		}
		return nil
	})
	var wg sync.WaitGroup
	sem := make(chan struct{}, internal.Config.General.TranslateStreams) // limit to internal.Config.General.TranslateStreams
	var mu sync.Mutex
	for _, t := range toTranslate {
		wg.Add(1)
		sem <- struct{}{}
		go func(t toTranslateT) {
			defer wg.Done()
			defer func() { <-sem }()
			translation, err := api.Translate(t.translate)
			if err != nil {
				log.Printf("Error translating '%s' from %s to %s: %s", t.translate.Text, t.translate.From, t.translate.To, err.Error())
				TryAgainTranslation(t.translate.From, t.translate.To, t.translate.Text)
			} else {
				SaveTranslation(t.translate.From, t.translate.To, t.translate.Text, translation)
			}
			keyExists := []byte(translationsDeferredPrefix + t.translate.From + "_" + t.translate.To + "_" + helpers.Md5Hash(t.translate.Text))
			mu.Lock()
			defer mu.Unlock()
			_ = bdb.Update(func(txn *badger.Txn) error {
				_ = txn.Delete(t.key)
				_ = txn.Delete(keyExists)
				return nil
			})
		}(t)
	}

	wg.Wait()
}
