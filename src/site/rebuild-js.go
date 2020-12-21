package site

import (
	"errors"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var rebuildJSMutex sync.Mutex

func RebuildJS(path string, config *Config) error {
	rebuildJSMutex.Lock()
	defer rebuildJSMutex.Unlock()
	var entryFiles = make([]string, 0, len(config.Javascript.Entries))
	for _, e := range config.Javascript.Entries {
		entryFile := filepath.Join(path, e)
		if _, err := os.Stat(entryFile); err != nil {
			err := errors.New(fmt.Sprintf("can't access entry file %s: %s", entryFile, err.Error()))
			return err
		}
		entryFiles = append(entryFiles, entryFile)
	}
	outDir := filepath.Join(path, "../public")
	if config.Javascript.Destination != "" {
		outDir = filepath.Join(outDir, config.Javascript.Destination)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		err := errors.New(fmt.Sprintf("can't create out directory for javascript bundle: %s", err.Error()))
		return err
	}
	result := api.Build(api.BuildOptions{
		EntryPoints:       entryFiles,
		Outdir:            outDir,
		Bundle:            true,
		Write:             true,
		LogLevel:          api.LogLevelInfo,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	})
	for _, m := range result.Errors {
		log.Println("Error in", m.Location.File, m.Location.Line, ":", m.Text)
	}
	for _, m := range result.Warnings {
		log.Println("Warning in", m.Location.File, m.Location.Line, ":", m.Text)
	}
	return nil
}
