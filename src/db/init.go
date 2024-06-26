package db

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"

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
			bdb, err = badger.Open(
				badger.DefaultOptions(internal.Config.Database.Path).
					WithDetectConflicts(false).
					WithValueLogFileSize(50 << 20). // 50 MB
					WithIndexCacheSize(10 << 20).   // 10 MB
					WithBlockCacheSize(10 << 20).   // 10 MB
					WithNumMemtables(2).
					WithSyncWrites(false).
					WithLoggingLevel(badger.WARNING),
			)
		} else {
			bdb, err = badger.Open(
				badger.LSMOnlyOptions(internal.Config.Database.Path).
					WithDetectConflicts(false).
					WithSyncWrites(false).
					// WithValueLogMaxEntries(100000).
					WithValueLogFileSize(250 << 20). // 250 MB
					WithIndexCacheSize(2000 << 20). // 2 GB
					WithBlockCacheSize(100 << 20).   // 100 MB
					WithNumMemtables(5).
					WithNumLevelZeroTables(1).
					WithNumLevelZeroTablesStall(2).
					WithLoggingLevel(badger.WARNING),
			)
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
	// Garbage collector
	go func() {
		var err error
		for {
			time.Sleep(time.Second*30 + time.Second*time.Duration(rand.Intn(60)))
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("recover in badger db maintenance", r)
					}
				}()
				err = bdb.RunValueLogGC(0.5)
				if err != nil && err != badger.ErrNoRewrite {
					log.Println(err)
				}
			}()
		}
	}()
	// Translations
	go func() {
		for {
			time.Sleep(time.Millisecond * 1000)
			doTranslations()
			time.Sleep(time.Millisecond*2000 + time.Millisecond*time.Duration(rand.Intn(3000)))
		}
	}()
}

func BeforeClose() {
	if bdb != nil {
		err := bdb.Close()
		if err != nil {
			log.Println(err)
		}
	}
}
