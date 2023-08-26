package internal

import (
	"log"
	"path/filepath"
	"sersh.com/totaltube/frontend/types"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/rjeczalik/notify"
)

var configsMap = make(map[string]*types.Config)
var configsMutex sync.RWMutex

func GetConfig(configPath string) *types.Config {
	configsMutex.RLock()
	defer configsMutex.RUnlock()
	if config, ok := configsMap[configPath]; ok {
		return config
	}
	return GetConfigAndWatch(configPath)
}

func GetConfigAndWatch(configPath string) *types.Config {
	var config = types.NewConfig()
	config.Hostname = filepath.Base(filepath.Dir(configPath))
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		log.Fatalln("error reading site config at", configPath, err)
	}
	configsMap[configPath] = config
	go func() {
		var m sync.Mutex
		var lastChange time.Time
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("error in config file watching routine: %+v", r)
						time.Sleep(time.Second * 30)
					}
				}()
				c := make(chan notify.EventInfo, 1)
				if err := notify.Watch(filepath.Dir(configPath), c, notify.Write); err != nil {
					log.Panicln(configPath, err)
				}
				defer notify.Stop(c)
				// waiting for signal of file changing
				for {
					info := <-c
					if filepath.Base(info.Path()) == "config.toml" {
						break
					}
				}
				m.Lock()
				lastChange = time.Now()
				m.Unlock()
				// 1.5 seconds after last change we reload the config
				go func() {
					time.Sleep(time.Millisecond * 1500)
					m.Lock()
					defer m.Unlock()
					if !lastChange.After(time.Now().Add(-time.Millisecond * 1500)) {
						// reload config
						lastChange = time.Now()
						var newConfig = types.NewConfig()
						newConfig.Hostname = filepath.Base(filepath.Dir(configPath))
						if _, err := toml.DecodeFile(configPath, newConfig); err != nil {
							log.Println("error reading site config at", configPath, err)
						} else {
							configsMutex.Lock()
							configsMap[configPath] = newConfig
							configsMutex.Unlock()
							log.Println("config file " + configPath + " reloaded")
						}
					}
				}()
			}()
		}
	}()
	return config
}
