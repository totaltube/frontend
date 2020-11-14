package db

import (
	"github.com/dgraph-io/badger/v2"
	"log"
	"math/rand"
	"sersh.com/totaltube/frontend/internal"
	"time"
)

var bdb *badger.DB

func InitDB() {
	rand.Seed(time.Now().UnixNano())
	var err error
	log.Println("Init DB")
	bdb, err = badger.Open(
		badger.DefaultOptions(internal.Config.Database.Path).
			WithDetectConflicts(false).
			WithSyncWrites(false),
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
				break
			}
		}
	}()
}

func BeforeClose() {
	if bdb != nil {
		_ = bdb.Close()
	}
}
