package db

import (
	"github.com/dgraph-io/badger/v2"
	"log"
	"math/rand"
	"runtime"
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
			WithTruncate(runtime.GOOS == "windows"),
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
}

func BeforeClose() {
	if bdb != nil {
		err := bdb.Close()
		if err != nil {
			log.Println(err)
		}
	}
}
