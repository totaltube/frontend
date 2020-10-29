package main

import (
	"fmt"
	_ "github.com/gofiber/fiber/v2"
	"github.com/jpillora/overseer"
	"log"
	"runtime"
	"sersh.com/totaltube/frontend/internal"
)

var version = "dev"
var serverPort = 1000

func main() {
	if runtime.GOOS == "windows" {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetFlags(log.Lshortfile)
	}
	internal.Version = version
	if runtime.GOOS == "windows" { // под виндой не работает overseer
		server(overseer.State{
			GracefulShutdown: make(chan bool),
		})
	} else {
		overseer.Run(overseer.Config{
			Program: server,
			Address: fmt.Sprintf(":%d", serverPort),
			Debug:   false,
		})
	}
}
