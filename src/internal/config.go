package internal

import (
	"github.com/BurntSushi/toml"
	"log"
	"sersh.com/totaltube/frontend/types"
)

var Config *ConfigT

type (
	ConfigT struct {
		General  General
		Frontend Frontend
		Database Database
	}
	General struct {
		Nginx          bool `toml:"nginx"`
		Port           uint
		RealIpHeader   string         `toml:"real_ip_header"`
		UseIpV6Network bool           `toml:"use_ipv6_network"`
		ApiUrl         string         `toml:"api_url"`
		ApiSecret      string         `toml:"api_secret"`
		ApiTimeout     types.Duration `toml:"api_timeout"`
		LangCookie     string         `toml:"lang_cookie"`
	}
	Frontend struct {
		SitesPath   string `toml:"sites_path"`
		DefaultSite string `toml:"default_site"`
		SecretKey   string `toml:"secret_key"`
	}
	Database struct {
		Path string `toml:"path"`
	}
)

func InitConfig(configPath string) {
	Config = &ConfigT{General: General{Nginx: true}}
	if _, err := toml.DecodeFile(configPath, Config); err != nil {
		log.Fatalln(configPath, ":", err)
	}
}
