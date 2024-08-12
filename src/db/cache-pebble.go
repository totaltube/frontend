package db

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cockroachdb/pebble"

	"sersh.com/totaltube/frontend/helpers"
)

func recreateJobPebble(job recreateInfo) {
	defer func() {
		if job.doneChannel != nil {
			defer func() {
				close(job.doneChannel)
			}()
		}
		if r := recover(); r != nil {
			log.Println("recover in cache recreate worker: ", r)
			debug.PrintStack()
			if job.doneChannel != nil {
				job.doneChannel <- fmt.Errorf("%s", r)
			}
		}
	}()
	var keyStr = cachePrefix + job.cacheKey
	var key = []byte(keyStr)
	var expireKeyStr = cachePrefix + "_exp_" + job.cacheKey
	var expireKey = []byte(expireKeyStr)
	defer recreatingNow.Delete(job.cacheKey)
	result, err := job.recreateFunction()
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Println(err)
		}
	} else {
		err = pdb.Set(key, result, pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
		// we set ttl slightly higher than requested timeout, because we want to use old cache sometimes
		err = pdb.Set([]byte("ie_"+time.Now().Add(job.timeout+job.extendedTimeout).Format(time.RFC3339)), []byte(keyStr), pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
		err = pdb.Set(expireKey, []byte(time.Now().Add(job.timeout).Format(time.RFC3339)), pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
		err = pdb.Set([]byte("ie_"+time.Now().Add(job.timeout+job.extendedTimeout).Format(time.RFC3339)), []byte(expireKeyStr), pebble.NoSync)
		if err != nil {
			log.Println(err)
			return
		}
	}
	if job.doneChannel != nil {
		job.doneChannel <- err
	}
}

func GetCachedTimeoutPebble(cacheKey string, timeout time.Duration, extendedTimeout time.Duration, recreate func() ([]byte, error), bypassCache bool) (result []byte, err error) {
	key := []byte(cachePrefix + cacheKey)
	expireKey := []byte(cachePrefix + "_exp_" + cacheKey)
	found := false
	var expire time.Time
	value, closer, err := pdb.Get(key)
	if err == nil {
		// found cached value
		result = bytes.Clone(value)
		_ = closer.Close()
		value, closer, err = pdb.Get(expireKey)
		if err == nil {
			expire, err = time.Parse(time.RFC3339, string(value))
			_ = closer.Close()
			if err != nil {
				log.Println(err)
				return
			}
			found = true
		}
	} else if !errors.Is(err, pebble.ErrNotFound) {
		log.Println(err)
		return
	}
	if found && !bypassCache {
		if expire.After(time.Now()) { // there are some time for expiration, just return found cached value
			return
		}
		if extendedTimeout == 0 {
			// recreating in place
			result, err = recreate()
			if err != nil {
				log.Println(err)
				return
			}
			// Saving to cache
			err = pdb.Set(key, bytes.Clone(result), pebble.NoSync)
			if err != nil {
				log.Println(err)
				return
			}
			err = pdb.Set([]byte("ie_"+time.Now().Add(timeout+extendedTimeout).Format(time.RFC3339)+"_"+helpers.RandStr(10)), key, pebble.NoSync)
			if err != nil {
				log.Println(err)
				return
			}
			err = pdb.Set(expireKey, []byte(time.Now().Add(timeout).Format(time.RFC3339)), pebble.NoSync)
			if err != nil {
				log.Println(err)
				return
			}
			err = pdb.Set([]byte("ie_"+time.Now().Add(timeout+extendedTimeout).Format(time.RFC3339)+"_"+helpers.RandStr(10)), expireKey, pebble.NoSync)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
		// need to recreate found cached value. But let's check if we already have recreate job for this
		if _, loaded := recreatingNow.LoadOrStore(cacheKey, time.Now()); loaded {
			// currently we recreating value, so, just return old one cached
			return
		}
		// recreating in background
		var info = recreateInfo{
			recreateFunction: recreate,
			cacheKey:         cacheKey,
			timeout:          timeout,
			extendedTimeout:  extendedTimeout,
		}
		if strings.HasPrefix(cacheKey, "in:") {
			innerRecreateQueue <- info
		} else {
			recreateQueue <- info
		}
		return
	}
	// if not found cached value, just recreate it right now, but let's check if we already trying this and wait for another thread to done with it
	waited := false
	for {
		if _, loaded := recreatingNow.LoadOrStore(cacheKey, time.Now()); !loaded {
			break
		}
		waited = true
		time.Sleep(time.Millisecond * 10)
	}
	if waited {
		// trying to get recreated by another thread cached value
		recreatingNow.Delete(cacheKey)
		return GetCachedTimeout(cacheKey, timeout, extendedTimeout, recreate, false)
	}
	// recreating
	var done = make(chan error)
	var info = recreateInfo{
		recreateFunction: recreate,
		cacheKey:         cacheKey,
		timeout:          timeout,
		extendedTimeout:  extendedTimeout,
		doneChannel:      done,
	}
	startTime := time.Now()
	if strings.HasPrefix(cacheKey, "in:") {
		innerRecreateQueue <- info
	} else {
		recreateQueue <- info
	}
	err = <-done
	elapsed := time.Since(startTime)
	if elapsed > time.Second*5 {
		log.Println(cacheKey, "too long time to recreate: ", elapsed, err)
	}
	if err != nil {
		// Removing old cache
		if found {
			_ = pdb.Delete(key, pebble.NoSync)
			_ = pdb.Delete(expireKey, pebble.NoSync)
		}
		return
	}
	return GetCachedTimeoutPebble(cacheKey, timeout, extendedTimeout, recreate, false)
}

func deleteKeysPebble(keysForDelete [][]byte) error {
	b := pdb.NewBatch()
	defer b.Close()
	for _, key := range keysForDelete {
		if err := b.Delete(key, pebble.NoSync); err != nil {
			return err
		}
	}
	if err := b.Commit(pebble.NoSync); err != nil {
		return err
	}
	return nil
}

func ClearCacheByPrefixPebble(prefix string) (err error) {
	var Prefix = []byte(cachePrefix + prefix)
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()

	collectSize := 10000
	var it *pebble.Iterator
	it, err = pdb.NewIter(&pebble.IterOptions{
		LowerBound: Prefix,
	})
	if err != nil {
		return
	}
	defer it.Close()
	keysForDelete := make([][]byte, 0, collectSize)
	keysCollected := 0
	for valid := it.First(); valid; valid = it.Next() {
		key := bytes.Clone(it.Key())
		if !bytes.HasPrefix(key, Prefix) {
			break
		}
		keysForDelete = append(keysForDelete, key)
		keysCollected++
		if keysCollected >= collectSize {
			if err := deleteKeysPebble(keysForDelete); err != nil {
				return err
			}
			keysForDelete = make([][]byte, 0, collectSize)
			keysCollected = 0
		}
	}
	if keysCollected > 0 {
		if err := deleteKeysPebble(keysForDelete); err != nil {
			return err
		}
	}
	return
}
