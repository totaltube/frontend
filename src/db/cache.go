package db

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"

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
				job.doneChannel <- fmt.Errorf("%s", r)
			}
		}
	}()
	var key = []byte(cachePrefix + job.cacheKey)
	var expireKey = []byte(cachePrefix + "_exp_" + job.cacheKey)
	defer recreatingNow.Delete(job.cacheKey)
	result, err := job.recreateFunction()
	if err != nil {
		if !strings.Contains(err.Error(), "not found") && err.Error() != "custom response" {
			log.Println(err)
		}
	} else {
		err = bdb.Update(func(txn *badger.Txn) (err error) {
			// we set ttl slightly higher than requested timeout, because we want to use old cache sometimes
			entry := badger.NewEntry(key, result).WithTTL(job.timeout + job.extendedTimeout)
			err = txn.SetEntry(entry)
			if err != nil {
				return
			}
			expireEntry := badger.NewEntry(expireKey, []byte(time.Now().Add(job.timeout).Format(time.RFC3339))).
				WithTTL(job.timeout + job.extendedTimeout)
			err = txn.SetEntry(expireEntry)
			return
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

type cacheUpdate struct {
	done chan struct{}
}

var cacheUpdates sync.Map

func readFromBadgerCache(key, expireKey []byte) (data []byte, found bool, expired bool, err error) {
	var expire time.Time

	err = bdb.View(func(txn *badger.Txn) error {
		item, e := txn.Get(key)
		if e == badger.ErrKeyNotFound {
			return e // ключ не найден
		}
		if e != nil {
			return e // любая другая ошибка Badger
		}
		val, e2 := item.ValueCopy(nil)
		if e2 != nil {
			return e2
		}
		data = val

		// Читаем expireKey, если он есть
		itemExp, e3 := txn.Get(expireKey)
		if e3 == badger.ErrKeyNotFound {
			// значит, TTL не ставили
			return nil
		}
		if e3 != nil {
			return e3
		}
		expBytes, e4 := itemExp.ValueCopy(nil)
		if e4 != nil {
			return e4
		}
		t, e5 := time.Parse(time.RFC3339, string(expBytes))
		if e5 == nil {
			expire = t
		}
		return nil
	})
	if err == badger.ErrKeyNotFound {
		// Ключа (или expireKey) нет в базе
		return nil, false, false, nil
	}
	if err != nil {
		// Любая другая ошибка
		return nil, false, false, err
	}

	// Если дошли сюда, значит данные найдены
	found = true
	// Проверяем, не просрочены ли
	if !expire.IsZero() && time.Now().After(expire) {
		expired = true
	}
	return data, found, expired, nil
}

// storeToBadgerCache открывает Update-транзакцию и записывает данные с учётом TTL.
// timeout + extendedTime = реальная длительность хранения в Badger.
// Кроме того, в expireKey записываем дату "предварительного" окончания (timeout),
// чтобы понимать, когда данные «начнут считаться устаревшими» внутри приложения.
func storeToBadgerCache(key, expireKey []byte, data []byte, timeout, extendedTime time.Duration) error {
	return bdb.Update(func(txn *badger.Txn) error {
		// TTL в Badger будет timeout + extendedTime
		ttl := timeout + extendedTime

		entry := badger.NewEntry(key, data).WithTTL(ttl)
		if err := txn.SetEntry(entry); err != nil {
			return err
		}

		expireTime := time.Now().Add(timeout) // Когда данные «протухнут» для нашего кода
		expEntry := badger.NewEntry(expireKey, []byte(expireTime.Format(time.RFC3339))).
			WithTTL(ttl)

		return txn.SetEntry(expEntry)
	})
}

func GetCachedTimeout(
	cacheKey string,
	timeout, extendedTimeout time.Duration,
	recreate func() ([]byte, error),
	bypassCache bool,
) (result []byte, err error) {

	key := []byte(cachePrefix + cacheKey)
	expireKey := []byte(cachePrefix + "_exp_" + cacheKey)

	// 1. Проверяем, не занята ли уже кем-то реконструкция кэша
	updatePtr, loaded := cacheUpdates.LoadOrStore(cacheKey, &cacheUpdate{done: make(chan struct{})})
	update := updatePtr.(*cacheUpdate)

	if loaded {
		// Значит, другая горутина занимается обновлением этого ключа
		<-update.done // Ждём, пока она закончит

		// Затем пытаемся прочитать из кэша (или пересоздать, если нет)
		if !bypassCache {
			data, found, expired, err := readFromBadgerCache(key, expireKey)
			if err != nil {
				return nil, err
			}
			if found && !expired {
				// У нас есть валидные данные — сразу возвращаем
				return data, nil
			}
		}
		// Если ничего нет — пересоздаем
		return recreateAndStore(cacheKey, timeout, extendedTimeout, recreate)
	}

	// Мы "первые" — берём на себя обновление
	defer func() {
		// В любом случае по выходу разблокируем других
		close(update.done)
		cacheUpdates.Delete(cacheKey)
	}()

	// 2. Сразу читаем из кэша, если bypassCache = false
	if !bypassCache {
		data, found, expired, _ := readFromBadgerCache(key, expireKey)
		if found && !expired {
			// Кэш актуален
			return data, nil
		}
		// Иначе надо пересоздавать (либо не найден, либо просрочен)
	}

	// 3. Пересоздаём и записываем в кэш
	return recreateAndStore(cacheKey, timeout, extendedTimeout, recreate)
}

func recreateAndStore(
	cacheKey string,
	timeout, extendedTimeout time.Duration,
	recreate func() ([]byte, error),
) ([]byte, error) {
	result, err := recreate()
	if err != nil {
		return nil, err
	}
	key := []byte(cachePrefix + cacheKey)
	expireKey := []byte(cachePrefix + "_exp_" + cacheKey)

	// Записываем в Badger
	storeErr := storeToBadgerCache(key, expireKey, result, timeout, extendedTimeout)
	if storeErr != nil {
		log.Println("error storing to cache:", storeErr)
	}
	return result, err
}

var clearCacheMutex sync.Mutex

func ClearCacheByPrefix(prefix string) (err error) {
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()

	var (
		Prefix      = []byte(cachePrefix + prefix)
		collectSize = 10000
	)

	opts := badger.DefaultIteratorOptions
	opts.AllVersions = false
	opts.PrefetchValues = false

	startKey := Prefix
	for {
		var keysToDelete [][]byte

		// Read up to 10,000 keys in a View transaction
		err = bdb.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(opts)
			defer it.Close()

			it.Seek(startKey)
			count := 0
			for it.ValidForPrefix(Prefix) && count < collectSize {
				key := it.Item().KeyCopy(nil)
				keysToDelete = append(keysToDelete, key)
				count++
				it.Next()
			}
			return nil
		})
		if err != nil {
			return err
		}

		// If there are no keys, exit
		if len(keysToDelete) == 0 {
			break
		}

		// Delete collected keys
		err = bdb.Update(func(txn *badger.Txn) error {
			for _, key := range keysToDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		// If we did not collect a full "batch" of 10,000 keys,
		// it means the keys are finished — exit.
		if len(keysToDelete) < collectSize {
			break
		}

		// Otherwise, form a new startKey for the next iteration.
		// Since all keys are sorted, we can take the last deleted key
		// and add a 0x00 byte to it to guarantee moving forward.
		lastKey := keysToDelete[len(keysToDelete)-1]
		startKey = append(lastKey, 0x00)
	}

	return nil
}
