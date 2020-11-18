package site

import (
	"github.com/evanw/esbuild/pkg/api"
	"log"
	"os"
	"path/filepath"
)

func RebuildJS(path string, config *Config) {
	var entryFiles = make([]string, 0, len(config.Javascript.Entries))
	for _, e := range config.Javascript.Entries {
		entryFile := filepath.Join(path, e)
		if _, err := os.Stat(entryFile); err != nil {
			log.Println("can't access entry file", entryFile)
			return
		}
		entryFiles = append(entryFiles, entryFile)
	}

	result := api.Build(api.BuildOptions{
		EntryPoints: entryFiles,
		Outdir:     filepath.Join(path, "../public"),
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		MinifyWhitespace: true,
		MinifyIdentifiers: true,
		MinifySyntax: true,
	})
	for _, m := range result.Errors {
		log.Println("Error in", m.Location.File, m.Location.Line, ":", m.Text)
	}
	for _, m := range result.Warnings {
		log.Println("Warning in", m.Location.File, m.Location.Line, ":", m.Text)
	}
}
