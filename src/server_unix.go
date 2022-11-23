// +build linux darwin freebsd

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/jpillora/overseer"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/handlers"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/site"
)

func startServer() {
	overseer.Run(overseer.Config{
		Program: server,
		Address: fmt.Sprintf(":%d", internal.Config.General.Port),
		Debug:   false,
	})
}

func server(state overseer.State) {
	pidFile := filepath.Join(internal.Config.MainPath, "totaltube-frontend.pid")
	if data, err := ioutil.ReadFile(pidFile); err == nil {
		// there is some pid file
		if pid, err := strconv.ParseInt(string(data), 10, 32); err != nil && pid > 0 {
			for {
				if exists, _ := PidExists(int32(pid)); exists {
					// waiting until existing program exits
					time.Sleep(time.Millisecond * 300)
					continue
				}
				break
			}
		}
	}
	// writing new pid
	pid := os.Getppid()
	if err := ioutil.WriteFile(pidFile, []byte(strconv.FormatInt(int64(pid), 10)), 0640); err != nil {
		log.Fatalln("Can't save pid file:", err)
	}
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
		err := http.Serve(state.Listener, app)
		if err != nil {
			fmt.Println(err)
		}
	}()
	/*c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGABRT)
	signal.Notify(c, syscall.SIGKILL)
	signal.Notify(c, syscall.SIGUSR2)*/
	select {
	//case <-c:
	case <-state.GracefulShutdown:
	}
	// The program is going to finish
	log.Println("Making some cleanup before exit...")
	db.BeforeClose()
	// Deleting pid file
	if err := os.Remove(pidFile); err != nil {
		log.Println("can't remove pid file: ", err)
	}
}

func PidExists(pid int32) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid pid %v", pid)
	}
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}
