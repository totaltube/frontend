package db

import (
	"bytes"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/dgraph-io/badger/v4"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
)

var pdb *pebble.DB

func InitPebble() {
	c := pebble.NewCache(16384 << 20) // 16 GB
	defer c.Unref()
	for {
		var err error
		pdb, err = pebble.Open(filepath.Join(internal.Config.Database.Path, "pebble"), &pebble.Options{
			//MemTableSize: 2048 << 20, // 2 GB
			//BytesPerSync: 5 << 20,    // 5 MB
			Cache: c,
		})
		if err != nil {
			if strings.Contains(err.Error(), "temporarily unavailable") {
				time.Sleep(time.Millisecond * 200)
				continue
			}
			log.Fatalln("Can't open pebble database:", err)
		}
		break
	}
	if internal.Config.Database.RestoreFromBackup {
		go restoreFromBackupPebble()
	}
	// Translations
	go func() {
		for {
			time.Sleep(time.Millisecond * 1000)
			doTranslationsPebble()
			time.Sleep(time.Millisecond*2000 + time.Millisecond*time.Duration(rand.Intn(3000)))
		}
	}()
	// expired items cleanup
	go func() {
		for {
			time.Sleep(time.Second*30 + time.Second*time.Duration(rand.Intn(100)))
			// cleanup expired items
			cleanupExpiredItemsPebble()
		}
	}()
}

func restoreFromBackupPebble() {
	if _, err := os.Stat(filepath.Join(internal.Config.Database.BackupPath, "pebble.backup")); err != nil {
		log.Println("No backup pebble database found, will try to restore from badger")
		if _, err := os.Stat(filepath.Join(internal.Config.Database.BackupPath, "current.backup")); err != nil {
			log.Println("No badger backup file found. Nothing to restore")
			return
		}
		// first - we create temporary badger db and restore where from backup
		os.MkdirAll(filepath.Join(internal.Config.Database.BackupPath, "temp.badger"), 0755)
		var tempBadger *badger.DB
		tempBadger, err = badger.Open(badger.DefaultOptions(filepath.Join(internal.Config.Database.BackupPath, "temp.badger")).
			WithNumMemtables(1).
			WithNumLevelZeroTables(1).
			WithNumLevelZeroTablesStall(2).
			WithValueThreshold(10 << 10), // 10 KB
		)
		if err != nil {
			log.Println("Can't open temporary badger database:", err)
			return
		}
		defer func() {
			tempBadger.Close()
			os.RemoveAll(filepath.Join(internal.Config.Database.BackupPath, "temp.badger"))
		}()
		// restore from backup
		var file *os.File
		file, err = os.Open(filepath.Join(internal.Config.Database.BackupPath, "current.backup"))
		if err != nil {
			log.Println(err)
			return
		}
		err = tempBadger.Load(file, 16)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Restored temp badgerdb from backup file", file.Name(), ". Now restoring to pebble")
		// now we copy data from badger to bolt
		ch := make(chan [2][]byte)
		keysPerTransaction := 1000
		restored := 0
		restore := func(data [][2][]byte) {
			b := pdb.NewBatch()
			defer b.Close()
			for _, kv := range data {
				if err := b.Set(kv[0], kv[1], nil); err != nil {
					log.Println("Can't set key-value to pebble:", err)
					return
				}
				if bytes.HasPrefix(kv[0], []byte(translationsPrefix)) {
					expireKey := []byte("ie_" + time.Now().Add(expireAccessedTranslationsInterval).Format(time.RFC3339) + "_" + helpers.RandStr(10))
					err = b.Set(expireKey, kv[0], nil)
				} else if bytes.HasPrefix(kv[0], []byte(translationAccessedPrefix)) {
					expireKey := []byte("ie_" + time.Now().Add(updateAccessedTranslationsInterval).Format(time.RFC3339) + "_" + helpers.RandStr(10))
					err = b.Set(expireKey, kv[0], nil)
				} else if bytes.HasPrefix(kv[0], []byte(translationsDeferredPrefix)) {
					expireKey := []byte("ie_" + time.Now().Add(time.Hour).Format(time.RFC3339) + "_" + helpers.RandStr(10))
					err = b.Set(expireKey, kv[0], nil)
				} else if bytes.HasPrefix(kv[0], []byte(translationsTriedPrefix)) {
					expireKey := []byte("ie_" + time.Now().Add(time.Hour).Format(time.RFC3339) + "_" + helpers.RandStr(10))
					err = b.Set(expireKey, kv[0], nil)
				}
				if err != nil {
					log.Println("Can't set expire key to pebble:", err)
					return
				}
			}
			if err := b.Commit(nil); err != nil {
				log.Println("Can't commit batch to pebble:", err)
				return
			}
		}
		for i := 0; i < 3; i++ {
			go func() {
				// restore to pebble
				var data [][2][]byte
				for kv := range ch {
					data = append(data, kv)
					restored += 1
					if restored%100000 == 0 {
						log.Println("Restored", restored, "items from badger backup")
					}
					if len(data) < keysPerTransaction {
						continue
					}
					restore(data)
					data = nil
				}
				if len(data) > 0 {
					restore(data)
				}
			}()
		}
		_ = tempBadger.View(func(txn *badger.Txn) error {
			// create cursor and iterate over all keys
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				v, err := item.ValueCopy(nil)
				if err != nil {
					log.Println("Can't copy value from badger:", err)
					continue
				}
				ch <- [2][]byte{k, v}
			}
			return nil
		})
		close(ch)
		log.Println("Restored from badger backup")
		return
	}
}

var regexExpTime = regexp.MustCompile(`ie_([^_]+)`)

func cleanupExpiredItemsPebble() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in cleanupExpiredItemsPebble", r)
		}
	}()
	iter, err := pdb.NewIter(&pebble.IterOptions{
		LowerBound: []byte("ie_"),
	})
	if err != nil {
		log.Println("Error creating iterator for expired items cleanup:", err)
		return
	}
	defer iter.Close()

	now := time.Now()
	const batchSizeLimit = 1000 // Limit for the number of keys in one batch
	var keysToDelete [][]byte   // Slice to accumulate keys

	for valid := iter.SeekGE([]byte("ie_")); valid; valid = iter.Next() {
		key := bytes.Clone(iter.Key())
		if key == nil {
			break
		}
		if !bytes.HasPrefix(key, []byte("ie_")) {
			break
		}
		originalKey := bytes.Clone(iter.Value())
		keyParts := string(key)
		result := regexExpTime.FindStringSubmatch(keyParts)
		if result == nil {
			log.Println("Error parsing expiration time for key", string(key))
			continue
		}
		expireAt, err := time.Parse(time.RFC3339, result[1])
		if err != nil {
			log.Println("Error parsing expiration time for key", string(key), err)
			continue
		}

		if expireAt.Before(now) {
			keysToDelete = append(keysToDelete, key, originalKey) // Add keys to the slice

			// If the limit is reached, send the keys for deletion
			if len(keysToDelete) >= batchSizeLimit*2 {
				executeBatchDelete(keysToDelete)
				keysToDelete = nil                // Clear the slice after the batch is executed
				time.Sleep(time.Millisecond * 10) // Sleep for a while to avoid high CPU usage
			}
		} else {
			break
		}
	}

	// Delete remaining keys, if any
	if len(keysToDelete) > 0 {
		executeBatchDelete(keysToDelete)
	}
}

func executeBatchDelete(keys [][]byte) {
	batch := pdb.NewBatch()
	defer batch.Close()

	for _, key := range keys {
		if err := batch.Delete(key, nil); err != nil {
			log.Println("Error adding key to batch for deletion", string(key), err)
		}
	}

	// Commit the batch after adding all keys
	if err := batch.Commit(pebble.Sync); err != nil {
		log.Println("Error committing batch for expired items cleanup:", err)
	}
}

func BeforeClosePebble() {
	if pdb != nil {
		pdb.Close()
	}
}
