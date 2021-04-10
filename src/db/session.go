package db

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/segmentio/encoding/json"
	"log"
	"sersh.com/totaltube/frontend/helpers"
	"time"
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

func GetSession(ip string, lock ...bool) (session *Session) {
	if len(lock) > 0 && lock[0] {
		helpers.KeyMutex.Lock(ip)
	}
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

func SaveSession(ip string, session *Session, unlock ...bool) {
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
		return txn.Set(key, val)
	})
	if len(unlock) > 0 && unlock[0] {
		helpers.KeyMutex.Unlock(ip)
	}
}
