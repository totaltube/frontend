package db

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/cockroachdb/pebble"

	"sersh.com/totaltube/frontend/helpers"
)

func GetSessionPebble(ip string) (session *Session) {
	helpers.KeyMutex.Lock(ip)
	key := []byte(sessionPrefix + ip)
	value, closer, err := pdb.Get(key)
	if err == nil {
		defer closer.Close()
		session = new(Session)
		err = json.Unmarshal(bytes.Clone(value), session)
		_ = closer.Close()
		if err != nil {
			log.Println(err)
			session = nil
		}
	}
	if session == nil {
		session = new(Session)
		session.Ip = ip
	}
	return
}

func SaveSessionPebble(ip string, session *Session) {
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
	err = pdb.Set(key, val, pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
	err = pdb.Set([]byte("ie_"+time.Now().Add(time.Hour*4).Format(time.RFC3339)+"_"+helpers.RandStr(10)), key, pebble.NoSync)
	if err != nil {
		log.Println(err)
		return
	}
}
