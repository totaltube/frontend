package site

import "log"

func RebuildSCSS(path string, config *Config) {
	log.Println(path, config.Scss.Entries)
}
