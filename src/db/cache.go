package db

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"

	"sersh.com/totaltube/frontend/internal"
)

const (
	cachePrefix = "c_"
)

type recreateInfo struct {
	recreateFunction func() ([]byte, error)
	cacheKey         string
	timeout          time.Duration
	extendedTimeout  time.Duration
	doneChannel      chan error
}

var recreatingNow sync.Map
var recreateQueue chan recreateInfo
var innerRecreateQueue chan recreateInfo // extra queue for requests inside recreate functions to avoid deadlock

func recreateJob(job recreateInfo) {
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
	var key = []byte(cachePrefix + job.cacheKey)
	var expireKey = []byte(cachePrefix + "_exp_" + job.cacheKey)
	defer recreatingNow.Delete(job.cacheKey)
	result, err := job.recreateFunction()
	if err != nil  {
		if !strings.Contains(err.Error(), "not found") {
			log.Println(err)
		}
	} else {
		err = bdb.Update(func(txn *badger.Txn) error {
			// we set ttl slightly higher than requested timeout, because we want to use old cache sometimes
			entry := badger.NewEntry(key, result).WithTTL(job.timeout + job.extendedTimeout)
			err := txn.SetEntry(entry)
			if err != nil {
				return err
			}
			expireEntry := badger.NewEntry(expireKey, []byte(time.Now().Add(job.timeout).Format(time.RFC3339))).
				WithTTL(job.timeout + job.extendedTimeout)
			err = txn.SetEntry(expireEntry)
			return err
		})
	}
	if job.doneChannel != nil {
		job.doneChannel <- err
	}
}

func recreateWorker() {
	for job := range recreateQueue {
		recreateJob(job)
	}
}
func innerRecreateWorker() {
	for job := range innerRecreateQueue {
		recreateJob(job)
	}
}

func launchCacheWorkers() {
	recreateQueue = make(chan recreateInfo, internal.Config.General.RecreateWorkers*10)
	innerRecreateQueue = make(chan recreateInfo, internal.Config.General.InnerRecreateWorkers+10) // extra queue for requests inside recreate functions to avoid deadlock
	for i := 0; i < int(internal.Config.General.RecreateWorkers); i++ {
		go recreateWorker()
	}
	for i := 0; i < int(internal.Config.General.InnerRecreateWorkers); i++ {
		go innerRecreateWorker()
	}
}

func GetCachedTimeout(cacheKey string, timeout time.Duration, extendedTimeout time.Duration, recreate func() ([]byte, error), bypassCache bool) (result []byte, err error) {
	key := []byte(cachePrefix + cacheKey)
	expireKey := []byte(cachePrefix + "_exp_" + cacheKey)
	found := false
	var expire time.Time
	_ = bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		result, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		item, err = txn.Get(expireKey)
		if err != nil {
			return err
		}
		var expireBytes []byte
		expireBytes, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		expire, err = time.Parse(time.RFC3339, string(expireBytes))
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
	elapsed := time.Now().Sub(startTime)
	if elapsed > time.Second*5 {
		log.Println(cacheKey, "too long time to recreate: ", elapsed, err)
	}
	if err != nil {
		// Removing old cache
		if found {
			_ = bdb.Update(func(txn *badger.Txn) error {
				_ = txn.Delete(key)
				_ = txn.Delete(expireKey)
				return nil
			})
		}
		return
	}
	return GetCachedTimeout(cacheKey, timeout, extendedTimeout, recreate, false)
}

var clearCacheMutex sync.Mutex

func ClearCacheByPrefix(prefix string) (err error) {
	var Prefix = []byte(cachePrefix + prefix)
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := bdb.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	collectSize := 10000
	err = bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(Prefix); it.ValidForPrefix(Prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					it.Close()
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		it.Close()
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}
		return nil
	})
	return
}
