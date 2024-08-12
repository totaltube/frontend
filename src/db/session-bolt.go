package db

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
	"sersh.com/totaltube/frontend/helpers"
)

func GetSessionBolt(ip string) (session *Session) {
	helpers.KeyMutex.Lock(ip)
	key := []byte(ip)
	err := boltdb.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("sessions")).Get(key)
		if v == nil {
			return nil
		}
		session = new(Session)
		decompressed, err := Decompress(v)
		if err != nil {
			return err
		}
		err = json.Unmarshal(decompressed, session)
		return err
	})
	if err != nil {
		log.Println(err)
	}
	if session == nil {
		session = new(Session)
		session.Ip = ip
	}
	return
}

func SaveSessionBolt(ip string, session *Session) {
	defer helpers.KeyMutex.Unlock(ip)
	if session == nil {
		return
	}
	v, err := json.Marshal(session)
	if err != nil {
		log.Println(err)
		return
	}
	key := []byte(ip)
	err = boltdb.Update(func(tx *bbolt.Tx) (err error) {
		err = tx.Bucket([]byte("sessions")).Put(key, Compress(v))
		if err != nil {
			return
		}
		err = tx.Bucket([]byte("sessions_expire")).Put(key, []byte(time.Now().Add(time.Hour*4).Format(time.RFC3339)))
		return
	})
	if err != nil {
		log.Println(err)
	}
}
