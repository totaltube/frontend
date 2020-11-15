package db

import (
	"github.com/dgraph-io/badger/v2"
	"time"
)

const (
	cachePrefix = "c_"
)

func GetCached(cacheKey string) []byte {
	key := []byte(cachePrefix + cacheKey)
	var cached []byte
	_ = bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		cached, err = item.ValueCopy(nil)
		return err
	})
	return cached
}

func PutCached(cacheKey string, content []byte, ttl time.Duration) (err error) {
	key := []byte(cachePrefix + cacheKey)
	err = bdb.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, content).WithTTL(ttl)
		err := txn.SetEntry(entry)
		return err
	})
	return
}
