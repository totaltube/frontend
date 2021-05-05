package db

import (
	"github.com/dgraph-io/badger/v3"
	"sync"
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

var clearCacheMutex sync.Mutex
func ClearCacheByPrefix(prefix string) (err error) {
	var Prefix = []byte(cachePrefix + prefix)
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := bdb.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	collectSize := 10000
	err = bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(Prefix); it.ValidForPrefix(Prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					it.Close()
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		it.Close()
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}
		return nil
	})
	return
}