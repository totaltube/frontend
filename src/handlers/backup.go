package handlers

import (
	"net/http"
)

var Backup = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Backup the database
	//err := db.DoBackup("backup.db")

})
