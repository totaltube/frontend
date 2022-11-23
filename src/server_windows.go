// +build windows

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jpillora/overseer"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/handlers"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
)

func startServer() {
	server(overseer.State{
		GracefulShutdown: make(chan bool),
	})
}
func server(_ overseer.State) {
	log.Println("Initializing database...")
	db.InitDB()
	log.Println("Initializing languages...")
	initLanguages()
	log.Println("Initializing pongo templates...")
	site.InitPongo2()
	log.Println("Initializing backgrounds...")
	handlers.InitBackgrounds()
	log.Println("Initializing minifier...")
	helpers.InitMinifier()
	log.Println("Initializing router...")
	app := InitRouter()
	go func() {
		log.Println("Running totaltube-frontend on port", internal.Config.General.Port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", internal.Config.General.Port), app)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(app.Listen(fmt.Sprintf(":%d", internal.Config.General.Port)))
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGABRT)
	signal.Notify(c, syscall.SIGKILL)
	<-c
	// The program is going to finish
	log.Println("Making some cleanup before exit...")
	db.BeforeClose()
}
