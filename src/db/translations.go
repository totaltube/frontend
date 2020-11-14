package db

import (
	"github.com/dgraph-io/badger/v2"
	"log"
	"math/rand"
	"sersh.com/totaltube/frontend/helpers"
	"strconv"
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
		translation = item.String()
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
			// тогда добавляем рандом к ключу
			key = append(key, []byte(strconv.FormatInt(rand.Int63n(100500000), 10))...)
		}
		_ = txn.SetEntry(badger.NewEntry(triedKey, []byte(now)).WithTTL(time.Minute * 60)) // Минимум раз в час пробуем еще раз перевести
		_ = txn.SetEntry(badger.NewEntry(key, helpers.ToJSON(translationDoc{
			From: from,
			To:   to,
			Text: text,
		})).WithTTL(time.Minute * 60))
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
