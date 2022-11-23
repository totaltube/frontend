package helpers

import (
	"log"
	"sync"
	"time"
)

type KeyMutexT struct {
	mu sync.Map
}
var empty struct {}

func (k *KeyMutexT) Lock(key interface{}) {
	for {
		if _, loaded := k.mu.LoadOrStore(key, empty); loaded {
			// waiting for releasing of lock by key
			time.Sleep(time.Millisecond*5)
		} else {
			break
		}
	}
}

func (k *KeyMutexT) Unlock(key interface{}) {
	if _, loaded := k.mu.LoadAndDelete(key); !loaded {
		log.Println("lock with key", key, "already released!")
	}
}

var KeyMutex = new(KeyMutexT)
