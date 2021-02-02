package db

import (
	"github.com/dgraph-io/badger/v3"
	"log"
	"math/rand"
	"sersh.com/totaltube/frontend/internal"
	"time"
)

var bdb *badger.DB

func InitDB() {
	rand.Seed(time.Now().UnixNano())
	var err error
	bdb, err = badger.Open(
		badger.DefaultOptions(internal.Config.Database.Path).
			WithDetectConflicts(false).
			WithSyncWrites(false).
			WithLoggingLevel(badger.ERROR),
	)
	if err != nil {
		log.Fatalln(err)
	}
	// Garbage collector
	go func() {
		var err error
		for {
			time.Sleep(time.Second*60 + time.Second*time.Duration(rand.Intn(60)))
			err = bdb.RunValueLogGC(0.7)
			if err != nil && err != badger.ErrNoRewrite {
				log.Println(err)
			}
		}
	}()
	// Translations
	go func() {
		for {
			time.Sleep(time.Millisecond*1000)
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
