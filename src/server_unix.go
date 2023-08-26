//go:build linux || darwin || freebsd

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/geoip"
	"sersh.com/totaltube/frontend/handlers"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/site"
	"strconv"
	"sync"
	"syscall"
	"time"

	"sersh.com/totaltube/frontend/internal"
)

var child *exec.Cmd
var mu sync.Mutex
var reloading = true

func startServer() {
	/*overseer.Run(overseer.Config{
		Program: server,
		Address: fmt.Sprintf(":%d", internal.Config.General.Port),
		Debug:   false,
	})*/
	pidFile := filepath.Join(internal.Config.MainPath, "totaltube-frontend.pid")
	if data, err := os.ReadFile(pidFile); err == nil {
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
	if err := os.WriteFile(pidFile, []byte(strconv.FormatInt(int64(pid), 10)), 0640); err != nil {
		log.Fatalln("Can't save pid file:", err)
	}
	log.Printf("Starting master process on port %d. Master process PID is %d\n", internal.Config.General.Port, os.Getpid())
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", internal.Config.General.Port))
	if err != nil {
		log.Printf("Error listening on %d: %s\n", internal.Config.General.Port, err)
		os.Exit(1)
	}
	defer listener.Close()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGQUIT)

	go func() {
		for {
			sig := <-sigs
			log.Println("Got signal", sig)
			if sig == syscall.SIGUSR2 || sig == syscall.SIGHUP {
				log.Println("Reloading worker...")
				go reload()
				continue
			}
			mu.Lock()
			if child != nil {
				err := child.Process.Signal(syscall.SIGTERM)
				if err != nil {
					log.Println(err)
				}
			}
			mu.Unlock()
			os.Exit(0)
		}
	}()
	go reload()
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting client: %s\n", err)
		}
		var childPid int
		for {
			mu.Lock()
			if reloading || child == nil {
				mu.Unlock()
				time.Sleep(time.Millisecond * 5)
				continue
			}
			childPid = child.Process.Pid
			mu.Unlock()
			break
		}
		go handleRequest(client, childPid)
	}
}
func handleRequest(src net.Conn, pid int) {
	defer src.Close()
	socketFile := filepath.Join(os.TempDir(), fmt.Sprintf("totaltube-frontend-%d.sock", pid))
	dst, err := net.Dial("unix", socketFile)
	if err != nil {
		log.Printf("Error connecting to unix socket: %s", err.Error())
		mu.Lock()
		defer mu.Unlock()
		if !reloading {
			if syscall.Kill(pid, syscall.Signal(0)) != nil {
				reloading = true
				go reload()
			}
		}
		return
	}
	defer dst.Close()

	// Send data from src to dst and dst to src
	go io.Copy(src, dst)
	io.Copy(dst, src)
}

func reload() {
	// Send signal to child to stop
	// We might need to track the PIDs of child processes
	// syscall.Kill(pid, syscall.SIGINT)
	mu.Lock()
	reloading = true
	oldChild := child
	mu.Unlock()
	// Start new child process
	args := make([]string, 0, 10)
	args = append(args, "child")
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "start" {
			continue
		}
		args = append(args, os.Args[i])
	}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalln("can't run worker:", err)
	}
	if cmd.Process == nil {
		log.Fatalln("can't run worker:", err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	if oldChild != nil {
		err = oldChild.Process.Signal(syscall.SIGINT)
		if err != nil {
			log.Println(err)
		}
	}
	// waiting until new process will begin listening
	for {
		select {
		case err = <-done:
			if err != nil {
				log.Fatalf("worker process exited with error: %v\n", err)
			}
			log.Fatalln("worker process exited unexpectedly")
			return
		default:
			// Continue to wait for child listening
		}
		socketFile := filepath.Join(os.TempDir(), fmt.Sprintf("totaltube-frontend-%d.sock", cmd.Process.Pid))
		if _, err := os.Stat(socketFile); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 5)
	}
	mu.Lock()
	child = cmd
	reloading = false
	mu.Unlock()
	go func() {
		err := <-done
		if err != nil {
			log.Printf("worker process finished with error: %s\n", err)
		}
		log.Println("Old worker process finished")
	}()
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

func startChild() {
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
	log.Println("Initializing country groups...")
	initCountryGroups()
	log.Println("Initializing GeoIP...")
	geoip.InitGeoIP(internal.Config.Database.Path, internal.Config.General.GeoipUrl)
	log.Println("Initializing router...")
	app := InitRouter()
	socketFile := filepath.Join(os.TempDir(), fmt.Sprintf("totaltube-frontend-%d.sock", os.Getpid()))
	listener, err := net.Listen("unix", socketFile)
	if err != nil {
		log.Fatalf("Failed to listen on unix socket: %v", err)
	}
	defer listener.Close()
	srv := &http.Server{
		Handler: app,
	}
	go func() {
		log.Printf("Worker with PID %d started\n", os.Getpid())
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			// Unexpected server error
			fmt.Printf("ListenAndServe(): %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	s := <-stop
	if s == syscall.SIGINT {
		// Finishing the server with timeout, continue to handle current requests
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error shutting down the server
			fmt.Printf("Server Shutdown: %v", err)
		}
	}
	// The program is going to finish
	log.Println("Making some cleanup before exit...")
	db.BeforeClose()
}
