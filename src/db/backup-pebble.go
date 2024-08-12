package db

import (
	"sersh.com/totaltube/frontend/internal"
)

func doBackupsPebble() {
	if internal.Config.Database.BackupPath == "" {
		return
	}
}
