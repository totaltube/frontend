package site

import (
	"github.com/rjeczalik/notify"
	"log"
	"runtime"
	"sync"
	"time"
)

func WatchScss(path string, config *Config) {
	var rebuildTimeout = time.Millisecond * 1500
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		rebuildTimeout = time.Millisecond*100
	}
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
					time.Sleep(rebuildTimeout)
					mu.Lock()
					if !lastChange.After(time.Now().Add(-rebuildTimeout)) {
						lastChange = time.Now()
						mu.Unlock()
						log.Println(ei.Path(),  "changed. Rebuilding scss...")
						started := time.Now()
						err := RebuildSCSS(path, config)
						if err != nil {
							log.Println("Error rebuilding scss:", err)
						}
						log.Println("done rebuilding scss in", time.Now().Sub(started).Truncate(time.Millisecond))
					} else {
						mu.Unlock()
					}
				}()
			}()
		}
	}()
}
