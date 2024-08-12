package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
	"sersh.com/totaltube/frontend/internal"
)

func doBackupsBolt() {
	if internal.Config.Database.BackupPath == "" {
		return
	}
	err := os.MkdirAll(filepath.Join(internal.Config.Database.BackupPath), 0755)
	if err != nil {
		log.Println(err)
		return
	}
	var skip = false
	_ = boltdb.Update(func(tx *bbolt.Tx) error {
		if v := tx.Bucket([]byte("last")).Get([]byte("last_backup")); v != nil {
			skip = true
			return nil
		}
		_ = tx.Bucket([]byte("last")).Put([]byte("last_backup"), []byte("1"))
		_ = tx.Bucket([]byte("last_expire")).Put([]byte("last_backup"), []byte(time.Now().Format(time.RFC3339)))
		return nil
	})
	if skip {
		return
	}
	// First - we open temporary boltdb to copy only translations data
	var tempBoltdb *bbolt.DB
	tempBoltdb, err = bbolt.Open(filepath.Join(internal.Config.Database.BackupPath, "temp.bolt.db"), 0600, nil)
	if err != nil {
		log.Println("Can't open temporary bolt database:", err)
		return
	}
	defer func() {
		tempBoltdb.Close()
		os.Remove(filepath.Join(internal.Config.Database.BackupPath, "temp.bolt.db"))
	}()
	// create buckets
	err = tempBoltdb.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range []string{
			"translations", "translations_expire",
			"translations_access", "translations_access_expire",
			"translations_tried", "translations_tried_expire",
			"translations_deferred", "translations_deferred_expire",
		} {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println("Can't create temporary bolt database buckets:", err)
		return
	}
	// Copy data
	err = boltdb.View(func(tx *bbolt.Tx) error {
		for _, bucket := range []string{
			"translations", "translations_expire",
			"translations_access", "translations_access_expire",
			"translations_tried", "translations_tried_expire",
			"translations_deferred", "translations_deferred_expire",
		} {
			cursor := tx.Bucket([]byte(bucket)).Cursor()
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				_ = tempBoltdb.Batch(func(tx *bbolt.Tx) error {
					return tx.Bucket([]byte(bucket)).Put(k, v)
				})
			}
		}
		return nil
	})
	if err != nil {
		log.Println("Can't copy data to temporary bolt database:", err)
		return
	}
	tempBoltdb.Close()
	// move temp to final
	err = os.Rename(filepath.Join(internal.Config.Database.BackupPath, "temp.bolt.db"), filepath.Join(internal.Config.Database.BackupPath, "bolt.db"))
	if err != nil {
		log.Println("Can't move temporary bolt database to final:", err)
		return
	}
	log.Println("Backup of bold database is done")
}
