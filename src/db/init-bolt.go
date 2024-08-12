package db

import (
	"bytes"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/golang/snappy"
	"go.etcd.io/bbolt"
	"sersh.com/totaltube/frontend/internal"
)

var boltdb *bbolt.DB

func InitBolt() {
	var err error
	boltdb, err = bbolt.Open(filepath.Join(internal.Config.Database.Path, "bolt.db"), 0600, &bbolt.Options{
		FreelistType:   bbolt.FreelistMapType,
		NoFreelistSync: false,
	})
	if err != nil {
		log.Fatalln("Can't open bolt database:", err)
	}
	// create buckets
	err = boltdb.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range []string{
			"sessions", "sessions_expire",
			"cache", "cache_expire",
			"cache_extended", "cache_extended_expire",
			"translations", "translations_expire",
			"translations_access", "translations_access_expire",
			"translations_tried", "translations_tried_expire",
			"translations_deferred", "translations_deferred_expire",
			"last", "last_expire",
		} {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalln("Can't create bolt database buckets:", err)
	}
	if internal.Config.Database.RestoreFromBackup {
		go restoreFromBackupBolt()
	}
	// expired items cleanup
	go func() {
		for {
			time.Sleep(time.Second*300 + time.Second*time.Duration(rand.Intn(100)))
			for _, bucket := range []string{
				"cache", "sessions", "translations", "translations_access",
				"translations_tried", "translations_deferred", "cache_extended", "last",
			} {
				log.Println("cleaning expired items from bolt bucket", bucket)
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Println("recover in bolt db maintenance", r)
						}
					}()
					for {
						deletedCount := 0
						err := boltdb.Update(func(tx *bbolt.Tx) error {
							cursor := tx.Bucket([]byte(bucket + "_expire")).Cursor()
							from := time.Now().Format(time.RFC3339)
							for k, v := cursor.Seek([]byte(from)); k != nil; k, v = cursor.Next() {
								tx.Bucket([]byte(bucket)).Delete(v)
								cursor.Delete()
								deletedCount++
								if deletedCount > 100 {
									// doing it in small chunks
									break
								}
							}
							return nil
						})
						if err != nil {
							log.Println("Error cleaning expired items from bolt db:", err)
						}
						if deletedCount == 0 {
							// nothing to delete anymore
							break
						}
					}
				}()
			}
			log.Println("bolt db cleaned")
		}
	}()
}

func Compress(data []byte) []byte {
	return snappy.Encode(nil, data)
}

func Decompress(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}

func restoreFromBackupBolt() {
	// restore from backup
	if _, err := os.Stat(filepath.Join(internal.Config.Database.BackupPath, "bolt.db")); err != nil {
		log.Println("No backup bolt database found, will try to restore from badger")
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
		log.Println("Restored temp badgerdb from backup file", file.Name())
		// now we copy data from badger to bolt
		ch := make(chan [2][]byte)
		boltdb.MaxBatchDelay = 10 * time.Millisecond
		boltdb.MaxBatchSize = 100000
		keysPerTransaction := 5000
		restored := 0
		restore := func(data [][2][]byte) {
			_ = boltdb.Batch(func(tx *bbolt.Tx) error {
				for _, kv := range data {
					k, v := kv[0], kv[1]
					if bytes.HasPrefix(k, []byte(translationsPrefix)) {
						k := bytes.TrimPrefix(k, []byte(translationsPrefix))
						_ = tx.Bucket([]byte("translations")).Put(k, Compress(v))
						_ = tx.Bucket([]byte("translations_expire")).Put(k, []byte(time.Now().Add(expireAccessedTranslationsInterval).Format(time.RFC3339)))
					} else if bytes.HasPrefix(k, []byte(translationsDeferredPrefix)) {
						k = bytes.TrimPrefix(k, []byte(translationsDeferredPrefix))
						_ = tx.Bucket([]byte("translations_deferred")).Put(k, Compress(v))
						_ = tx.Bucket([]byte("translations_deferred_expire")).Put(k, []byte(time.Now().Add(time.Hour).Format(time.RFC3339)))
					} else if bytes.HasPrefix(k, []byte(translationsTriedPrefix)) {
						k = bytes.TrimPrefix(k, []byte(translationsTriedPrefix))
						_ = tx.Bucket([]byte("translations_tried")).Put(k, v)
						_ = tx.Bucket([]byte("translations_tried_expire")).Put(k, []byte(time.Now().Add(time.Hour).Format(time.RFC3339)))
					} else if bytes.HasPrefix(k, []byte(translationAccessedPrefix)) {
						k = bytes.TrimPrefix(k, []byte(translationAccessedPrefix))
						_ = tx.Bucket([]byte("translations_access")).Put(k, v)
						_ = tx.Bucket([]byte("translations_access_expire")).Put(k, []byte(time.Now().Add(updateAccessedTranslationsInterval).Format(time.RFC3339)))
					} else {
						log.Println("Unknown key prefix in badger backup:", string(k))
					}
				}
				return nil
			})

		}
		for i := 0; i < 30; i++ {
			go func() {
				// restore to bolt
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
	// restore from bolt backup
	backup, err := bbolt.Open(filepath.Join(internal.Config.Database.BackupPath, "bolt.db"), 0600, nil)
	if err != nil {
		log.Println("Can't open bolt backup database:", err)
		return
	}
	defer backup.Close()
	err = backup.View(func(tx *bbolt.Tx) error {
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			return b.ForEach(func(k, v []byte) error {
				_ = boltdb.Batch(func(tx *bbolt.Tx) error {
					_ = tx.Bucket(name).Put(k, v)
					return nil
				})
				return nil
			})
		})
		return nil
	})
	if err != nil {
		log.Println("Can't restore from bolt backup:", err)
		return
	}
	log.Println("Restored from bolt backup")
}
