package main

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dlclark/regexp2"

	"sersh.com/totaltube/frontend/internal"
)

func upgradeConfig() {
	matches, err := filepath.Glob(filepath.Join(internal.Config.Frontend.SitesPath, "*"))
	if err != nil {
		panic(err)
	}
	var regexRoute = regexp2.MustCompile(`\:([\w_]+)`, regexp2.None)
	var replaceFunc = func(match regexp2.Match) string {
		return "{" + match.Groups()[1].String() + "}"
	}
	for _, m := range matches {
		configPath := filepath.Join(m, "config.toml")
		if _, err := os.Stat(configPath); err != nil {
			continue
		}
		var bt []byte
		if bt, err = ioutil.ReadFile(configPath); err != nil {
			panic(err)
		}
		originalContent := string(bt)
		var updatedContent string
		updatedContent, err = regexRoute.ReplaceFunc(originalContent, replaceFunc, -1, -1)
		if err != nil {
			panic(err)
		}
		var jsMatches []string
		jsMatches, err = filepath.Glob(filepath.Join(m, "js", "*.ts"))
		if err != nil {
			panic(err)
		}
		for _, jsMatch := range jsMatches {
			jsMatch := jsMatch
			var bt []byte
			if bt, err = ioutil.ReadFile(jsMatch); err != nil {
				panic(err)
			}
			jsOriginal := string(bt)
			var updatedJs string
			updatedJs, err = regexRoute.ReplaceFunc(jsOriginal, replaceFunc, -1, -1)
			if err != nil {
				panic(err)
			}
			updatedJs = strings.ReplaceAll(updatedJs, "/\\.(\\d+)\\.jpg$/, \".%d.jpg\"", "/\\.(\\d+)\\.(webp|jpg|png)$/, \".%d.$2\"")
			updatedJs = strings.ReplaceAll(updatedJs, "/\\.jpg$/, \"@2x.jpg\"", "/\\.(jpg|webp|png)$/, \"@2x.$1\"")

			if updatedJs != jsOriginal {
				go func() {
					time.Sleep(time.Second * 5)
					err := ioutil.WriteFile(jsMatch, []byte(updatedJs), fs.ModePerm)
					if err != nil {
						log.Println(err)
					}
				}()
			}
		}
		if updatedContent != originalContent {
			err = ioutil.WriteFile(configPath, []byte(updatedContent), fs.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
