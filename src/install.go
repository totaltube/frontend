package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func Install() {
	var err error
	var mainPath string
	// asking for main path
	if runtime.GOOS == "windows" {
		userHomeDir, _ := os.UserHomeDir()
		mainPath = filepath.Join(userHomeDir, "totaltube-frontend")
	} else {
		mainPath = "/var/lib/totaltube-frontend"
	}
	err = survey.AskOne(&survey.Input{
		Message: "Main Path: ",
		Default: mainPath,
	}, &mainPath, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.MkdirAll(mainPath, 0750)
	if err != nil {
		fmt.Println(err)
		return
	}
	var config = internal.ConfigT{
		General: internal.General{
			Port:           8380,
			Nginx:          runtime.GOOS != "windows",
			RealIpHeader:   "X-Real-Ip",
			UseIpV6Network: true,
			ApiTimeout:     types.Duration(time.Second * 30),
			LangCookie:     "lang",
			Development:    runtime.GOOS == "windows",
		},
		Frontend: internal.Frontend{
			SitesPath:        filepath.Join(mainPath, "sites"),
			DefaultSite:      "",
			SecretKey:        helpers.RandStr(40),
			CaptchaKey:       "",
			MaxDmcaMinute:    3,
			CaptchaWhiteList: []string{},
		},
		Database: internal.Database{Path: filepath.Join(mainPath, "database")},
	}
	_, _ = toml.DecodeFile(filepath.Join(mainPath, "config.toml"), &config)
	// Asking for port
	err = survey.AskOne(&survey.Input{
		Message: "Port to listen on: ",
		Default: fmt.Sprintf("%d", config.General.Port),
	}, &config.General.Port, survey.WithValidator(func(ans interface{}) error {
		if p, err := strconv.ParseUint(ans.(string), 10, 16); err != nil {
			return err
		} else {
			ln, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", p), time.Second)
			if err == nil {
				_ = ln.Close()
				return errors.New("port " + ans.(string) + " is not free")
			}
		}
		return nil
	}))
	if err != nil {
		fmt.Println(err)
		return
	}
	// asking for real ip header
	err = survey.AskOne(&survey.Input{
		Message: "Real IP Header: ",
		Default: config.General.RealIpHeader,
	}, &config.General.RealIpHeader, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Println(err)
		return
	}
	// asking for api url
	err = survey.AskOne(&survey.Input{
		Message: "API URL",
		Help:    "Enter API URL of totaltube minion",
		Default: config.General.ApiUrl,
	}, &config.General.ApiUrl, survey.WithValidator(func(ans interface{}) error {
		if u, err := url.ParseRequestURI(ans.(string)); err != nil {
			return errors.Wrap(err, "Wrong URL")
		} else if u.Hostname() == "" {
			return errors.New("Wrong URL")
		}
		return nil
	}))
	if err != nil {
		fmt.Println(err)
		return
	}
	// Asking API secret
	err = survey.AskOne(&survey.Input{
		Message: "API secret",
		Default: config.General.ApiSecret,
	}, &config.General.ApiSecret, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Println(err)
		return
	}
	// Asking for captcha key
	err = survey.AskOne(&survey.Input{
		Message: "Captcha key (may be omitted now)",
		Default: config.Frontend.CaptchaKey,
	}, &config.Frontend.CaptchaKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Asking for captcha secret
	err = survey.AskOne(&survey.Input{
		Message: "Captcha secret (may be omitted now)",
		Default: config.Frontend.CaptchaSecret,
	}, &config.Frontend.CaptchaSecret)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = survey.AskOne(&survey.Input{
		Message: "Default site domain",
		Default: config.Frontend.DefaultSite,
	}, &config.Frontend.DefaultSite)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Creating config file and directories...")
	err = os.MkdirAll(filepath.Join(mainPath, "sites"), 0750)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.MkdirAll(filepath.Join(mainPath, "database"), 0750)
	if err != nil {
		fmt.Println(err)
		return
	}
	t := template.New("config")
	_, err = t.Parse(helpers.ConfigFileTemplate)
	if err != nil {
		log.Fatalln(err)
	}
	var cFile *os.File
	if cFile, err = os.Create(filepath.Join(mainPath, "config.toml")); err != nil {
		fmt.Println("can't create config file: ", err)
		return
	}
	defer cFile.Close()
	err = t.Execute(cFile, config)
	if err != nil {
		fmt.Println("can't write config file: ", err)
	}
	cFile.Close()
	switch runtime.GOOS {
	case "linux", "freebsd", "darwin":
		helpers.InstallBashCompletions()
		if runtime.GOOS == "freebsd" {
			err = FreeBSDInstall(mainPath)
		} else if _, e := os.Stat("/usr/lib/systemd"); !os.IsNotExist(e) {
			err = SystemdInstall(mainPath)
		} else {
			log.Fatalln("unknown startup system. Please run totaltube-frontend start manually and install it on startup. Ask support for workarounds.")
		}
		if err != nil {
			fmt.Println("error during installation: ", err)
			os.Exit(1)
		}
		fmt.Println("Totaltube frontend installed and running now.")
	case "windows":
		fmt.Println("done.")
	}
}


func FreeBSDInstall(path string) error {
	configPath := filepath.Join(path, "config.toml")
	cmd := exec.Command("pw", "user", "add", "totaltube", "-d", path,
		"-s", "/usr/sbin/nologin")
	_ = cmd.Run()
	_, err := user.Lookup("totaltube")
	if err != nil {
		return errors.New("somehow can't create totaltube user")
	}
	fmt.Println("user totaltube created")
	cmd = exec.Command("chown", "-R", "totaltube", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("can't chown " + path + ": " + string(out))
	}
	cmd = exec.Command("chmod", "0755", path)
	_ = cmd.Run()
	fmt.Println("installing Totaltube Frontend daemon...")
	f, err := os.Create("/etc/rc.d/totaltube-frontend")
	if err != nil {
		return errors.New("can't create file " + err.Error())
	}
	err = helpers.BSDInit(f, path, configPath)
	f.Close()
	if err != nil {
		return errors.New("can't write to file " + err.Error())
	}
	_ = os.Chmod("/etc/rc.d/totaltube-frontend", 0755)

	ffc, _ := ioutil.ReadFile("/etc/rc.conf")
	if !strings.Contains(string(ffc), "totaltube_enable") {
		ff, err := os.OpenFile("/etc/rc.conf", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return errors.New("can't open /etc/rc.conf: " + err.Error())
		}
		defer ff.Close()
		_, err = ff.WriteString("\ntotaltube_enable=\"YES\"\n")
		if err != nil {
			return errors.New("can't append to /etc/rc.conf: " + err.Error())
		}
	}
	fmt.Println("setting up logging...")
	_ = ioutil.WriteFile("/etc/syslog.d/totaltube-frontend.conf", []byte("!totaltube-frontend\n*.*     /var/log/totaltube-frontend.log"), 0644)
	_ = ioutil.WriteFile("/etc/newsyslog.conf.d/totaltube-frontend.conf", []byte("/var/log/totaltube-frontend.log                        644  7     1000  *     JC\n"), 0644)
	cmd = exec.Command("touch", "/var/log/totaltube-frontend.log")
	_ = cmd.Run()
	cmd = exec.Command("chown", "totaltube", "/var/log/totaltube-frontend.log")
	_ = cmd.Run()
	fmt.Println("reloading syslogd...")
	cmd = exec.Command("service", "syslogd", "reload")
	_ = cmd.Run()
	fmt.Println("starting totaltube frontend...")
	cmd = exec.Command("service", "totaltube-frontend", "start")
	_ = cmd.Start()
	return nil
}


func SystemdInstall(path string) error {
	configPath := filepath.Join(path, "config.toml")
	// Создаем юзера
	cmd := exec.Command("useradd", "-d", path, "-s", "/usr/sbin/nologin", "totaltube")
	_ = cmd.Run()
	cmd = exec.Command("chown", "-R", "totaltube", path)
	_ = cmd.Run()
	cmd = exec.Command("chmod", "0755", path)
	_ = cmd.Run()
	fp, err := os.Create("/etc/systemd/system/totaltube-frontend.service")
	if err != nil {
		return err
	}
	err = helpers.SystemdInit(fp, path, configPath)
	fp.Close()
	if err != nil {
		return err
	}
	_ = os.Chmod("/etc/systemd/system/totaltube-frontend.service", 0644)
	if _, err = os.Stat("/etc/rsyslog.d"); err == nil {
		_ = ioutil.WriteFile("/etc/rsyslog.d/totaltube-frontend.conf", []byte(`if $programname == 'totaltube-frontend' then /var/log/totaltube-frontend.log
& stop`), 0644)
		_ = ioutil.WriteFile("/etc/logrotate.d/totaltube-frontend.conf", []byte(`/var/log/totaltube-frontend.log
{
	rotate 10
	size=5M
	missingok
	copytruncate
	delaycompress
	nomail
	notifempty
	noolddir
	compress
}`), 0644)
	}
	cmd = exec.Command("systemctl", "restart", "rsyslog")
	_ = cmd.Run()
	cmd = exec.Command("systemctl", "restart", "rsyslogd")
	_ = cmd.Run()
	cmd = exec.Command("systemctl", "daemon-reload")
	_ = cmd.Run()
	cmd = exec.Command("systemctl", "enable", "totaltube-frontend")
	_ = cmd.Run()
	cmd = exec.Command("systemctl", "start", "totaltube-frontend")
	_ = cmd.Run()
	return nil
}