package internal

import (
	"errors"
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
	configsMutex.Lock()
	if config, ok := configsMap[configPath]; ok {
		configsMutex.Unlock()
		return config
	}
	configsMutex.Unlock()
	return GetConfigAndWatch(configPath, updateConfig)
}

func readConfig(configPath string) (config *types.Config, configSource string, err error) {
	config = types.NewConfig()
	config.Hostname = filepath.Base(filepath.Dir(configPath))
	configSourceBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("error reading config at", configPath, err)
		return
	}
	configSource = string(configSourceBytes)
	if _, err = toml.DecodeFile(configPath, config); err != nil {
		log.Println("error decoding config at", configPath, err)
		return
	}
	// прочитаем еще файл с окончанием -translations.toml
	// для этого найдем базовый путь без экстеншена и добавим к нему -translations.toml
	basePath := filepath.Dir(configPath)
	translationsPath := filepath.Join(basePath, "config-translations.toml")
	translationsConfig := types.ConfigTranslations{}
	translationsConfigAlternate := make(map[string]map[string]string)
	if _, errt := os.Stat(translationsPath); errt == nil {
		var err1, err2 error
		ok := false
		_, err1 = toml.DecodeFile(translationsPath, &translationsConfig)
		if err1 == nil && len(translationsConfig.Translations) > 0 {
			for k, trs := range translationsConfig.Translations {
				for k2, v2 := range trs {
					if _, ok := config.Translations[k]; !ok {
						config.Translations[k] = make(map[string]string)
					}
					config.Translations[k][k2] = v2
				}
			}
			ok = true
		} else {
			_, err2 = toml.DecodeFile(translationsPath, &translationsConfigAlternate)
			if err2 == nil && len(translationsConfigAlternate) > 0 {
				for k, trs := range translationsConfigAlternate {
					for k2, v2 := range trs {
						if _, ok := config.Translations[k]; !ok {
							config.Translations[k] = make(map[string]string)
						}
						config.Translations[k][k2] = v2
					}
				}
			}
			ok = true
		}
		if !ok && (err1 != nil || err2 != nil) {
			err = errors.Join(err1, err2)
			log.Println("error decoding translations config at", translationsPath, err)
		}
	}
	return
}

func GetConfigAndWatch(configPath string, updateConfig func(config *types.Config, configSource string) error) *types.Config {
	config, configSource, err := readConfig(configPath)
	if err != nil {
		log.Fatalln("error reading config source at", configPath, err)
	}
	go func() {
		if err := updateConfig(config, configSource); err != nil {
			log.Println("error updating config at", configPath, err)
		}
	}()

	for k, v := range Config.Custom {
		if _, ok := config.Custom[k]; !ok {
			config.Custom[k] = v
		}
	}
	configsMutex.Lock()
	configsMap[configPath] = config
	configsMutex.Unlock()
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
					if filepath.Base(info.Path()) == "config.toml" || filepath.Base(info.Path()) == "config-translations.toml" {
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
						newConfig, newConfigSource, err := readConfig(configPath)
						if err != nil {
							log.Println("error reading config source at", configPath, err)
							return
						}
						go func() {
							if err := updateConfig(newConfig, newConfigSource); err != nil {
								log.Println("error updating config at", configPath, err)
							}
						}()
						configsMutex.Lock()
						configsMap[configPath] = newConfig
						configsMutex.Unlock()
						log.Println("config file " + configPath + " reloaded")
					}
				}()
			}()
		}
	}()
	return config
}
