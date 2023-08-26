package site

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/wellington/go-libsass"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sersh.com/totaltube/frontend/types"
	"strings"
	"sync"
)

var rebuildSCSSMutex sync.Mutex
var cssMinifier *minify.M

func RebuildSCSS(path string, config *types.Config) error {
	rebuildSCSSMutex.Lock()
	defer rebuildSCSSMutex.Unlock()
	for _, entryBase := range config.Scss.Entries {
		entry := filepath.Join(path, entryBase)
		outName := strings.TrimSuffix(entryBase, filepath.Ext(entryBase)) + ".css"
		outDir := filepath.Join(path, "../public", config.Scss.Destination)
		outFile := filepath.Join(outDir, outName)
		err := func () error {
			reader, err := os.Open(entry)
			if err != nil {
				return errors.New(fmt.Sprintf("Can't open scss entry file %s: %s", entry, err.Error()))
			}
			defer reader.Close()
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return errors.New(fmt.Sprintf("Can't create out directory %s for compiled css: %s",
					outDir, err.Error()))
			}
			writer, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return errors.New(fmt.Sprintf("can't create file for compiled css output: %s", err.Error()))
			}
			defer writer.Close()
			compilation, err := libsass.New(writer, reader,
				libsass.IncludePaths([]string{path}),
				libsass.ImgDir(filepath.Join(path, config.Scss.ImagesPath)),
				libsass.ImgBuildDir(filepath.Join(outDir, config.Scss.ImagesPath)),
			)

			if err != nil {
				return errors.New(fmt.Sprintf("error compiling scss %s: %s", entry, err.Error()))
			}
			if err := compilation.Run(); err != nil {
				return errors.New(fmt.Sprintf("error compiling scss %s: %s", entry, err.Error()))
			}
			return nil
		}()
		if err != nil {
			log.Println(err)
			return err
		}
		if config.Scss.Minify {
			if cssMinifier == nil {
				cssMinifier = minify.New()
				cssMinifier.AddFunc("text/css", css.Minify)
			}
			fi, err := os.Stat(outFile)
			if err != nil {
				return errors.New("out file read error: "+ err.Error())
			}
			var outMinified = bytes.NewBuffer(make([]byte, 0, fi.Size()))
			r, _ := os.Open(outFile)
			if err := cssMinifier.Minify("text/css", outMinified, r); err != nil {
				r.Close()
				return errors.New("can't minify "+outFile+": "+err.Error())
			} else {
				err = ioutil.WriteFile(outFile, outMinified.Bytes(), 0644)
				if err != nil {
					return  errors.New("can't write minified version of css file: "+err.Error())
				}
			}
		}
	}
	return nil
}
