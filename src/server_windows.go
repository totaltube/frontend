// +build windows

package main

import (
	"fmt"
	"github.com/jpillora/overseer"
	"log"
	"os"
	"os/signal"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"syscall"
)

func server(_ overseer.State) {
	db.InitDB()
	initLanguages()
	site.InitFilters()
	app := InitFiber()
	go func() {
		log.Println("Running totaltube-frontend on port", internal.Config.General.Port)
		fmt.Println(app.Listen(fmt.Sprintf(":%d", internal.Config.General.Port)))
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGABRT)
	signal.Notify(c, syscall.SIGKILL)
	<-c
	// Здесь мы после завершения программы
	log.Println("Making some cleanup before exit...")
	db.BeforeClose()
}
