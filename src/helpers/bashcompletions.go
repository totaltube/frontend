package helpers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func InstallBashCompletions() {
	var bashConfFiles []string
	switch runtime.GOOS {
	case "darwin":
		bashConfFiles = []string{".bash_profile"}
	default:
		bashConfFiles = []string{".bashrc", ".bash_profile", ".bash_login", ".profile"}
	}
	home, _ := os.UserHomeDir()
	binPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}
	binPath, _ = filepath.Abs(binPath)
	completeString := fmt.Sprintf("complete -C %s totaltube-frontend", binPath)
	for _, rc := range bashConfFiles {
		if _, err := os.Stat(filepath.Join(home, rc)); err == nil {
			// config file exists. Checking if there is complete
			f, _ := os.Open(filepath.Join(home, rc))
			bt, _ := ioutil.ReadAll(f)
			f.Close()
			if !bytes.Contains(bt, []byte(completeString)) {
				// not yet, adding
				f, err = os.OpenFile(filepath.Join(home, rc), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
				if err != nil {
					log.Println("can't open file", filepath.Join(home, rc), "for writing:", err)
					continue
				}
				_, err = f.WriteString("\n" + completeString + "\n")
				if err != nil {
					log.Println("can't write to", filepath.Join(home, rc), ":", err)
				}
				f.Close()
			}
		}
	}
}
