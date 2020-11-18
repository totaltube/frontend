package site

import (
	"github.com/rjeczalik/notify"
	"log"
	"sync"
	"time"
)

func WatchScss(path string, config *Config) {
	go func() {
		mu := sync.Mutex{}
		lastChange := time.Now()
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("recovered in watch scss in path", path, r)
					}
				}()
				c := make(chan notify.EventInfo, 1)
				if err := notify.Watch(path+"/...", c, notify.All); err != nil {
					log.Panicln(err)
				}
				defer notify.Stop(c)
				// ждем сигнала при изменении файлов шаблонов
				ei := <-c
				mu.Lock()
				lastChange = time.Now()
				mu.Unlock()
				go func() {
					time.Sleep(time.Millisecond*1500)
					mu.Lock()
					if !lastChange.After(time.Now().Add(-time.Millisecond * 1500)) {
						lastChange = time.Now()
						mu.Unlock()
						log.Println(ei.Path(),  "changed. Rebuilding scss...")
						RebuildSCSS(path, config)
					} else {
						mu.Unlock()
					}
				}()
			}()
		}
	}()
}
