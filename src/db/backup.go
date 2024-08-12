package db

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"sersh.com/totaltube/frontend/internal"
)

var backupMutex sync.Mutex

func DoBackup(file *os.File, since uint64) (err error) {
	backupMutex.Lock()
	defer backupMutex.Unlock()
	// Do backup
	stream := bdb.NewStream()
	stream.NumGo = 2
	stream.LogPrefix = "Badger.Streaming"
	stream.Prefix = []byte(translationsPrefix)
	_, err = stream.Backup(file, since)
	return
}

func doBackups() {
	if internal.Config.Database.Engine == "pebble" {
		doBackupsPebble()
	}
	if internal.Config.Database.Engine == "bolt" {
		doBackupsBolt()
		return
	}
	if internal.Config.Database.BackupPath == "" {
		return
	}
	err := os.MkdirAll(internal.Config.Database.BackupPath, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	var skip = false
	_ = bdb.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("last_backup")); err == nil {
			skip = true
			return nil
		}
		_ = txn.SetEntry(badger.NewEntry([]byte("last_backup"), []byte("1")).WithTTL(time.Hour * 24))
		return nil
	})
	if skip {
		return
	}
	// Do backups
	var file *os.File
	file, err = os.Create(filepath.Join(internal.Config.Database.BackupPath, "current.backup.tmp"))
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	err = DoBackup(file, 0)
	if err != nil {
		log.Println(err)
		return
	}
	file.Close()
	// перемещает current.backup.tmp в current.backup
	err = os.Rename(filepath.Join(internal.Config.Database.BackupPath, "current.backup.tmp"), filepath.Join(internal.Config.Database.BackupPath, "current.backup"))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Backup of badger database is done")
}
