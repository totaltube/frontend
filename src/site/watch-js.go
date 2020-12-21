package site

import (
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func WatchJS(path string, config *Config) {
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
						log.Println("recovered in watch js in path", path, r)
					}
				}()
				c := make(chan notify.EventInfo, 1)
				if err := notify.Watch(path, c, notify.All); err != nil {
					log.Panicln(err)
				}
				matches, _ := filepath.Glob(filepath.Join(path, "*"))
				for _, m := range matches {
					if filepath.Base(m) == "node_modules" {
						continue
					}
					info, _ := os.Stat(m)
					if info.IsDir() {
						if err := notify.Watch(m+"/...", c, notify.All); err != nil {
							log.Panicln(err)
						}
					} else {
						if err := notify.Watch(m, c, notify.All); err != nil {
							log.Panicln(err)
						}
					}
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
						log.Println(ei.Path(),  "changed. Rebuilding js...")
						err := RebuildJS(path, config)
						if err != nil {
							log.Println(err)
						}
						log.Println("done rebuilding js.")
					} else {
						mu.Unlock()
					}
				}()
			}()
		}
	}()
}
