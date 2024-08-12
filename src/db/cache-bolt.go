package db

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

func recreateJobBolt(job recreateInfo) {
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
				job.doneChannel <- errors.New(fmt.Sprintf("%s", r))
			}
		}
	}()
	var key = []byte(job.cacheKey)
	defer recreatingNow.Delete(job.cacheKey)
	result, err := job.recreateFunction()
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Println(err)
		}
	} else {
		err = boltdb.Update(func(tx *bbolt.Tx) (err error) {
			err = tx.Bucket([]byte("cache")).Put(key, Compress(result))
			if err != nil {
				return
			}
			// we set ttl slightly higher than requested timeout, because we want to use old cache sometimes
			_ = tx.Bucket([]byte("cache_expire")).Put(key, []byte(time.Now().Add(job.timeout+job.extendedTimeout).Format(time.RFC3339)))
			_ = tx.Bucket([]byte("cache_extended")).Put(key, []byte(time.Now().Add(job.timeout).Format(time.RFC3339)))
			_ = tx.Bucket([]byte("cache_extended_expire")).Put(key, []byte(time.Now().Add(job.timeout+job.extendedTimeout).Format(time.RFC3339)))
			return
		})
	}
	if job.doneChannel != nil {
		job.doneChannel <- err
	}
}

func GetCachedTimeoutBolt(cacheKey string, timeout time.Duration, extendedTimeout time.Duration, recreate func() ([]byte, error), bypassCache bool) (result []byte, err error) {
	key := []byte(cacheKey)
	found := false
	var expire time.Time
	err = boltdb.View(func(tx *bbolt.Tx) (err error) {
		v := tx.Bucket([]byte("cache")).Get(key)
		if v == nil {
			return nil
		}
		result, err = Decompress(v)
		if err != nil {
			return err
		}
		v = tx.Bucket([]byte("cache_expire")).Get(key)
		if v == nil {
			return errors.New("expire not found")
		}
		expire, err = time.Parse(time.RFC3339, string(v))
		if err != nil {
			return err
		}
		found = true
		return nil
	})
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
			err = boltdb.Batch(func(tx *bbolt.Tx) error {
				err = tx.Bucket([]byte("cache")).Put(key, Compress(result))
				if err != nil {
					return err
				}
				_ = tx.Bucket([]byte("cache_expire")).Put(key, []byte(time.Now().Add(timeout+extendedTimeout).Format(time.RFC3339)))
				_ = tx.Bucket([]byte("cache_extended")).Put(key, []byte(time.Now().Add(timeout).Format(time.RFC3339)))
				_ = tx.Bucket([]byte("cache_extended_expire")).Put(key, []byte(time.Now().Add(timeout+extendedTimeout).Format(time.RFC3339)))
				return nil
			})
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
			_ = boltdb.Batch(func(tx *bbolt.Tx) error {
				_ = tx.Bucket([]byte("cache")).Delete(key)
				_ = tx.Bucket([]byte("cache_expire")).Delete(key)
				_ = tx.Bucket([]byte("cache_extended")).Delete(key)
				_ = tx.Bucket([]byte("cache_extended_expire")).Delete(key)
				return nil
			})
		}
		return
	}
	return GetCachedTimeoutBolt(cacheKey, timeout, extendedTimeout, recreate, false)
}

func deleteKeysBolt(keysForDelete [][]byte) error {
	err := boltdb.Batch(func(tx *bbolt.Tx) error {
		for _, key := range keysForDelete {
			_ = tx.Bucket([]byte("cache")).Delete(key)
			_ = tx.Bucket([]byte("cache_expire")).Delete(key)
			_ = tx.Bucket([]byte("cache_extended")).Delete(key)
			_ = tx.Bucket([]byte("cache_extended_expire")).Delete(key)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ClearCacheByPrefixBolt(prefix string) (err error) {
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()
	collectSize := 1000
	err = boltdb.View(func(tx *bbolt.Tx) error {
		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		cursor := tx.Bucket([]byte("cache")).Cursor()
		for k, _ := cursor.Seek([]byte("prefix")); k != nil && bytes.HasPrefix([]byte("prefix"), k); k, _ = cursor.Next() {
			keysForDelete = append(keysForDelete, k)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeysBolt(keysForDelete); err != nil {
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeysBolt(keysForDelete); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println("Error clearing cache by prefix:", err)
	}
	return
}
