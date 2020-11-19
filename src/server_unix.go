// +build linux darwin freebsd

package main

import (
	"fmt"
	"github.com/jpillora/overseer"
	"log"
	"os"
	"os/signal"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
	"syscall"
)

func startServer() {
	overseer.Run(overseer.Config{
		Program: server,
		Address: fmt.Sprintf(":%d", internal.Config.General.Port),
		Debug:   false,
	})
}
func server(state overseer.State) {
	db.InitDB()
	initLanguages()
	site.InitPongo2()
	helpers.InitMinifier()
	app := InitFiber()
	go func() {
		log.Println("Running totaltube-frontend on port", internal.Config.General.Port)
		fmt.Println(app.Listener(state.Listener))
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGABRT)
	signal.Notify(c, syscall.SIGKILL)
	signal.Notify(c, syscall.SIGUSR2)
	select {
	case <-c:
	case <-state.GracefulShutdown:
	}
	// Здесь мы после завершения программы
	log.Println("Making some cleanup before exit...")
	db.BeforeClose()
}
