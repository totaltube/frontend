package site

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/dlclark/regexp2"
	"github.com/rjeczalik/notify"
)
var configsMap = make(map[string]*Config)
var configsMutex sync.RWMutex


func GetConfig(configPath string) *Config {
	configsMutex.RLock()
	defer configsMutex.RUnlock()
	if config, ok := configsMap[configPath]; ok  {
		return config
	}
	return GetConfigAndWatch(configPath)
}

func GetConfigAndWatch(configPath string) *Config {
	var config = NewConfig()
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		log.Fatalln("error reading site config at", configPath, err)
	}
	compatRoutes(&config.Routes)
	configsMap[configPath] = config
	go func() {
		var m sync.Mutex
		var lastChange time.Time
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("error in config file watching routine: %+v", r)
						time.Sleep(time.Second*30)
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
						var newConfig = NewConfig()
						if _, err := toml.DecodeFile(configPath, newConfig); err != nil {
							log.Println("error reading site config at", configPath, err)
						} else {
							compatRoutes(&newConfig.Routes)
							configsMutex.Lock()
							configsMap[configPath] = newConfig
							configsMutex.Unlock()
							log.Println("config file "+configPath+" reloaded")
						}
					}
				}()
			}()
		}
	}()
	return config
}


var regexRoute = regexp2.MustCompile(`\:([\w_]+)`, regexp2.None)

func compatRoutes(routes *ConfigRoutes) {
	replaceFunc := func(match regexp2.Match) string {
		return "{"+match.Groups()[1].String()+"}"
	}
	routes.VideoEmbed, _ = regexRoute.ReplaceFunc(routes.VideoEmbed, replaceFunc, -1, -1)
	routes.Autocomplete, _ = regexRoute.ReplaceFunc(routes.Autocomplete, replaceFunc, -1, -1)
	routes.New, _ = regexRoute.ReplaceFunc(routes.New, replaceFunc, -1, -1)
	routes.Model, _ = regexRoute.ReplaceFunc(routes.Model, replaceFunc, -1, -1)
	routes.Dmca, _ = regexRoute.ReplaceFunc(routes.Dmca, replaceFunc, -1, -1)
	routes.Long, _ = regexRoute.ReplaceFunc(routes.Long, replaceFunc, -1, -1)
	routes.LanguageTemplate, _ = regexRoute.ReplaceFunc(routes.LanguageTemplate, replaceFunc, -1, -1)
	routes.Models, _ = regexRoute.ReplaceFunc(routes.Models, replaceFunc, -1, -1)
	routes.Channel, _ = regexRoute.ReplaceFunc(routes.Channel, replaceFunc, -1, -1)
	routes.Category, _ = regexRoute.ReplaceFunc(routes.Category, replaceFunc, -1, -1)
	routes.ContentItem, _ = regexRoute.ReplaceFunc(routes.ContentItem, replaceFunc, -1, -1)
	routes.FakePlayer, _ = regexRoute.ReplaceFunc(routes.FakePlayer, replaceFunc, -1, -1)
	routes.Popular, _ = regexRoute.ReplaceFunc(routes.Popular, replaceFunc, -1, -1)
	routes.Maintenance, _ = regexRoute.ReplaceFunc(routes.Maintenance, replaceFunc, -1, -1)
	routes.Out, _ = regexRoute.ReplaceFunc(routes.Out, replaceFunc, -1, -1)
	routes.Search, _ = regexRoute.ReplaceFunc(routes.Search, replaceFunc, -1, -1)
	routes.TopCategories, _ = regexRoute.ReplaceFunc(routes.TopCategories, replaceFunc, -1, -1)
	routes.TopContent, _ = regexRoute.ReplaceFunc(routes.TopContent, replaceFunc, -1, -1)
	if routes.Custom != nil {
		for k, rr := range routes.Custom {
			routes.Custom[k], _ = regexRoute.ReplaceFunc(rr, replaceFunc, -1, -1)
		}
	}
}
