package internal

import (
	"github.com/BurntSushi/toml"
	"log"
)

var Config *ConfigT

type (
	ConfigT struct {
		General General
	}
	General struct {
		Port           uint
		Secret         string
		RealIpHeader   string `toml:"real_ip_header"`
		UseIpV6Network bool   `toml:"use_ipv6_network"`
	}
)

func InitConfig(configPath string) {
	Config = &ConfigT{}
	if _, err := toml.DecodeFile(configPath, Config); err != nil {
		log.Fatalln(configPath, ":", err)
	}
}
