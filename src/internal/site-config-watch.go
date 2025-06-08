package internal

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sersh.com/totaltube/frontend/types"

	"github.com/BurntSushi/toml"
	"github.com/rjeczalik/notify"
)

var configsMap = make(map[string]*types.Config)
var configsMutex sync.RWMutex

func GetConfig(configPath string, updateConfig func(config *types.Config, configSource string) error) *types.Config {
	configsMutex.RLock()
	defer configsMutex.RUnlock()
	if config, ok := configsMap[configPath]; ok {
		return config
	}
	return GetConfigAndWatch(configPath, updateConfig)
}

func GetConfigAndWatch(configPath string, updateConfig func(config *types.Config, configSource string) error) *types.Config {
	var config = types.NewConfig()
	config.Hostname = filepath.Base(filepath.Dir(configPath))

	// читаем configSource из файла configPath
	configSourceBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln("error reading config source at", configPath, err)
	}
	configSource := string(configSourceBytes)
	go updateConfig(config, configSource)

	if _, err := toml.DecodeFile(configPath, config); err != nil {
		log.Fatalln("error reading site config at", configPath, err)
	}
	for k, v := range Config.Custom {
		if _, ok := config.Custom[k]; !ok {
			config.Custom[k] = v
		}
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
						for k, v := range Config.Custom {
							if _, ok := newConfig.Custom[k]; !ok {
								newConfig.Custom[k] = v
							}
						}
						// читаем configSource заново при перезагрузке
						newConfigSourceBytes, err := os.ReadFile(configPath)
						if err != nil {
							log.Println("error reading config source at", configPath, err)
						}
						newConfigSource := string(newConfigSourceBytes)
						go updateConfig(newConfig, newConfigSource)
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
