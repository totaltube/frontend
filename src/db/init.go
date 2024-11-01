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
	if internal.Config.Database.Engine == "pebble" {
		InitPebble()
		return
	}
	if internal.Config.Database.Engine == "bolt" {
		InitBolt()
		return
	}
	var err error
	for {
		if internal.Config.Database.LowMemory {
			// low memory options
			bdb, err = badger.Open(
				badger.DefaultOptions(internal.Config.Database.Path).
					WithDetectConflicts(false).
					WithValueLogFileSize(500 << 20). // 500 MB
					WithIndexCacheSize(50 << 20).    // 50 MB
					WithBlockCacheSize(10 << 20).    // 10 MB
					WithValueThreshold(10 << 10).    // 10 KB
					WithNumMemtables(2).
					WithSyncWrites(false).
					WithLoggingLevel(badger.WARNING).
					WithVerifyValueChecksum(true),
			)
		} else {
			bdb, err = badger.Open(
				badger.DefaultOptions(internal.Config.Database.Path).
					WithDetectConflicts(false).
					WithSyncWrites(false).
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
					WithVerifyValueChecksum(true),
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
			time.Sleep(time.Millisecond * 1000)
			doTranslations()
			time.Sleep(time.Millisecond*2000 + time.Millisecond*time.Duration(rand.Intn(3000)))
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
			//err := bdb.Flatten(10)
			/*prefixStats := make(map[string]struct {
				Count   int
				Expired int
				Size    int64
			})
			prefixStatsAll := make(map[string]struct {
				Count   int
				Expired int
				Size    int64
			})
			// Создаем транзакцию для чтения
			err := bdb.View(func(txn *badger.Txn) error {
				// Создаем итератор с опциями
				opts := badger.DefaultIteratorOptions
				opts.PrefetchValues = false // Чтобы не загружать значения сразу
				it := txn.NewIterator(opts)
				defer it.Close()
				for it.Rewind(); it.Valid(); it.Next() {
					item := it.Item()
					key := item.Key()
					expire := item.ExpiresAt()
					prefix := "_"
					// Находим префикс до первого символа '_'
					prefixEnd := strings.IndexByte(string(key), '_')
					if prefixEnd != -1 {
						prefix = string(key[:prefixEnd+1])
					}
					if expire == 0 || time.Unix(int64(expire), 0).After(time.Now().Add(time.Hour*2+time.Minute*30)) {
						// Обновляем статистику для префикса
						stats := prefixStats[prefix]
						stats.Count++
						stats.Size += int64(item.EstimatedSize())
						if item.IsDeletedOrExpired() {
							stats.Expired++
						}
						prefixStats[prefix] = stats
						//log.Println("Key", string(key), "expires at", time.Unix(int64(expire), 0).Format(time.RFC3339))
					}
					statsAll := prefixStatsAll[prefix]
					statsAll.Count++
					statsAll.Size += int64(item.EstimatedSize())
					if item.IsDeletedOrExpired() {
						statsAll.Expired++
					}
					prefixStatsAll[prefix] = statsAll
				}
				return nil
			})
			if err != nil {
				log.Println(err)
				return
			}
			// Выводим результаты
			for prefix, stats := range prefixStats {
				log.Printf("Prefix: %s, Count: %d, Expired: %d, Size: %.2f mbytes\n", prefix, stats.Count, stats.Expired, float64(stats.Size)/1024/1024)
			}
			log.Println("All keys:")
			for prefix, stats := range prefixStatsAll {
				log.Printf("Prefix: %s, Count: %d, Expired: %d, Size: %.2f mbytes\n", prefix, stats.Count, stats.Expired, float64(stats.Size)/1024/1024)
			}*/
		}()
	}
}

func BeforeClose() {
	if internal.Config.Database.Engine == "pebble" {
		BeforeClosePebble()
		return
	}
	if bdb != nil {
		err := bdb.Close()
		if err != nil {
			log.Println(err)
		}
	}
}
