package internal

import (
	"github.com/BurntSushi/toml"
	"log"
)

var Config *ConfigT

type (
	ConfigT struct {
		General  General
		Frontend Frontend
	}
	General struct {
		Port           uint
		RealIpHeader   string `toml:"real_ip_header"`
		UseIpV6Network bool   `toml:"use_ipv6_network"`
		ApiUrl         string `toml:"api_url"`
		ApiSecret      string `toml:"api_secret"`
	}
	Frontend struct {
		SitesPath   string `toml:"sites_path"`
		DefaultSite string `toml:"default_site"`
	}
)

func InitConfig(configPath string) {
	Config = &ConfigT{}
	if _, err := toml.DecodeFile(configPath, Config); err != nil {
		log.Fatalln(configPath, ":", err)
	}
}
