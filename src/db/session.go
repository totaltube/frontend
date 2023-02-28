package db

import (
	"encoding/json"
	"log"
	"time"

	"github.com/dgraph-io/badger/v3"

	"sersh.com/totaltube/frontend/helpers"
)

const (
	sessionPrefix = "s_"
)

type Session struct {
	Ip            string
	LastViewType  string
	LastViewId    int64
	LastClickType string
	LastClickId   int64
	LastSave      time.Time
	LastDmca      time.Time
	DmcaAmount    int64
}

func GetSession(ip string) (session *Session) {
	helpers.KeyMutex.Lock(ip)
	key := []byte(sessionPrefix + ip)
	_ = bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			session = new(Session)
			err := json.Unmarshal(val, session)
			return err
		})
		return err
	})
	if session == nil {
		session = new(Session)
		session.Ip = ip
	}
	return
}

func SaveSession(ip string, session *Session) {
	defer helpers.KeyMutex.Unlock(ip)
	if session == nil {
		return
	}
	session.LastSave = time.Now()
	key := []byte(sessionPrefix + ip)
	val, err := json.Marshal(session)
	if err != nil {
		log.Println(err)
		return
	}
	_ = bdb.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, val).WithTTL(time.Hour * 4)
		return txn.SetEntry(entry)
	})
}
