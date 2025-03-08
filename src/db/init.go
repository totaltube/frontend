package db

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
)

var bdb *badger.DB

func InitDB() {
	rand.Seed(time.Now().UnixNano())
	launchCacheWorkers()
	var err error
	for {
		if internal.Config.Database.LowMemory {
			// low memory options
			options := badger.DefaultOptions(internal.Config.Database.Path).
				WithDetectConflicts(internal.Config.Database.DetectConflicts).
				WithValueLogFileSize(500 << 20). // 500 MB
				WithIndexCacheSize(50 << 20).    // 50 MB
				WithBlockCacheSize(10 << 20).    // 10 MB
				WithValueThreshold(10 << 10).    // 10 KB
				WithNumMemtables(2).
				WithSyncWrites(internal.Config.Database.SyncWrites).
				WithLoggingLevel(badger.WARNING).
				WithVerifyValueChecksum(true)
			bdb, err = badger.Open(options)
		} else {
			options := badger.DefaultOptions(internal.Config.Database.Path).
				WithDetectConflicts(internal.Config.Database.DetectConflicts).
				WithSyncWrites(internal.Config.Database.SyncWrites).
				// WithValueLogMaxEntries(100000).
				//WithValueLogFileSize(250 << 20). // 250 MB
				WithIndexCacheSize(2000 << 20). // 2 GB
				//WithBlockCacheSize(100 << 20).   // 100 MB
				WithMemTableSize(1 << 20). // 1 MB
				WithNumMemtables(2).
				WithNumLevelZeroTables(1).
				WithNumLevelZeroTablesStall(2).
				//WithNumLevelZeroTablesStall(2).
				WithValueThreshold(10 << 10). // 10 KB
				WithLoggingLevel(badger.WARNING).
				WithVerifyValueChecksum(true)
			bdb, err = badger.Open(options)
		}
		if err != nil {
			// Waiting until not closed process will close the database.
			if strings.Contains(err.Error(), "Cannot acquire directory lock") {
				log.Println("waiting for database unlocking...")
				time.Sleep(time.Millisecond * 200)
				continue
			}
		}
		break
	}
	if err != nil {
		log.Fatalln("Badger DB initialization error:", err, "Try to remove files from db directory",
			internal.Config.Database.Path, "if nothing helps")
	}
	if internal.Config.Database.RestoreFromBackup {
		go func() {
			helpers.KeyMutex.Lock("db_operations_lock")
			defer helpers.KeyMutex.Unlock("db_operations_lock")
			var file *os.File
			file, err = os.Open(filepath.Join(internal.Config.Database.BackupPath, "current.backup"))
			if err != nil {
				log.Println(err)
			} else {
				err = bdb.Load(file, 16)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Restored from backup file", file.Name())
				}
			}
		}()
	}
	// Garbage collector
	go func() {
		for {
			time.Sleep(time.Second*30 + time.Second*time.Duration(rand.Intn(10)))
			func() {
				helpers.KeyMutex.Lock("db_operations_lock")
				defer helpers.KeyMutex.Unlock("db_operations_lock")
				defer func() {
					if r := recover(); r != nil {
						log.Println("recover in badger db maintenance", r)
					}
				}()
				// Запускаем GC до тех пор, пока он не вернёт ошибку badger.ErrNoRewrite
				for {
					err := bdb.RunValueLogGC(0.01)
					if err == badger.ErrNoRewrite {
						break
					}
					if err != nil {
						log.Println("Ошибка очистки badger: ", err)
						break
					}
				}
			}()
		}
	}()

	// Translations
	go func() {
		for {
			time.Sleep(time.Millisecond * 100)
			doTranslations()
			time.Sleep(time.Millisecond*100 + time.Millisecond*time.Duration(rand.Intn(100)))
		}
	}()
	// Backups
	go func() {
		for {
			time.Sleep(time.Millisecond * 1200)
			doBackups()
			time.Sleep(time.Second*60 + time.Second*time.Duration(rand.Intn(120)))
		}
	}()

	if internal.Config.Database.DebugBadger {
		go func() {
			log.Println("Badger DB debug started")
			log.Println("Histogram for all keys")
			bdb.PrintHistogram(nil)
			log.Println("Histogram for s_ prefix")
			bdb.PrintHistogram([]byte("s_"))
			log.Println("Histogram for c_ prefix")
			bdb.PrintHistogram([]byte("c_"))
			log.Println("Histogram for tr_ prefix")
			bdb.PrintHistogram([]byte("tr_"))
			log.Println("Histogram for tra_ prefix")
			bdb.PrintHistogram([]byte("tra_"))
			log.Println("Histogram for trd_ prefix")
			bdb.PrintHistogram([]byte("trd_"))
			log.Println("Histogram for trt_ prefix")
			bdb.PrintHistogram([]byte("trt_"))
			log.Println("Levels")
			helpers.DumpJSON(bdb.Levels())
			onDisk, uncompressed := bdb.EstimateSize([]byte("c_"))
			log.Println("Size of c_ prefix on disk:", onDisk, "uncompressed:", uncompressed)
		}()
	}
}

func BeforeClose() {
	if bdb != nil {
		err := bdb.Close()
		if err != nil {
			log.Println(err)
		}
	}
}
