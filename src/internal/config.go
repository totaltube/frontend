package internal

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"

	"sersh.com/totaltube/frontend/types"
)

var Config *ConfigT

type (
	ConfigT struct {
		MainPath string
		General  General
		Frontend Frontend
		Database Database
		Options  *Options
	}
	General struct {
		Nginx          bool `toml:"nginx"`
		Port           uint16
		RealIpHeader   string         `toml:"real_ip_header"`
		UseIpV6Network bool           `toml:"use_ipv6_network"`
		ApiUrl         string         `toml:"api_url"`
		ApiSecret      string         `toml:"api_secret"`
		ApiTimeout     types.Duration `toml:"api_timeout"`
		LangCookie     string         `toml:"lang_cookie"`
		GeoipUrl       string         `toml:"geoip_url"`
		Development    bool           `toml:"development"`
	}
	Frontend struct {
		SitesPath        string   `toml:"sites_path"`
		DefaultSite      string   `toml:"default_site"`
		SecretKey        string   `toml:"secret_key"`
		CaptchaKey       string   `toml:"captcha_key"`
		CaptchaSecret    string   `toml:"captcha_secret"`
		MaxDmcaMinute    int64    `toml:"max_dmca_minute"`
		CaptchaWhiteList []string `toml:"captcha_whitelist"`
	}
	Database struct {
		Path string `toml:"path"`
	}
)

func InitConfig(configPath string) {
	Config = &ConfigT{
		General: General{
			Nginx:       true,
			Development: runtime.GOOS == "windows",
			GeoipUrl:    "https://totaltraffictrader.com/geo/country.tar.gz",
		},
		Frontend: Frontend{
			MaxDmcaMinute: 5,
		},
	}
	if _, err := toml.DecodeFile(configPath, Config); err != nil {
		log.Fatalln(configPath, ":", err)
	}
	Config.MainPath = filepath.Dir(configPath)
}
