package internal

import (
	"log"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"

	"sersh.com/totaltube/frontend/types"
)

var Config *ConfigT

type (
	ConfigT struct {
		MainPath     string
		General      General
		Frontend     Frontend
		Database     Database
		Options      *Options
		Translations map[string]map[string]string `toml:"translations"`
	}
	General struct {
		Nginx                              bool `toml:"nginx"`
		Port                               uint16
		RealIpHeader                       string         `toml:"real_ip_header"`
		UseIpV6Network                     bool           `toml:"use_ipv6_network"`
		ApiUrl                             string         `toml:"api_url"`
		ApiSecret                          string         `toml:"api_secret"`
		ApiTimeout                         types.Duration `toml:"api_timeout"`
		LangCookie                         string         `toml:"lang_cookie"`
		RecreateWorkers                    uint16         `toml:"recreate_workers"`
		InnerRecreateWorkers               uint16         `toml:"inner_recreate_workers"`
		GeoipUrl                           string         `toml:"geoip_url"`
		Development                        bool           `toml:"development"`
		ToplistDataUrl                     string         `toml:"toplist_data_url"`
		DefaultBlackholeRoute              string         `toml:"default_blackhole_route"`
		CheckForBots                       bool           `toml:"check_for_bots"`
		EnableAccessLog                    bool           `toml:"enable_access_log"`
		DeletedTaxonomiesToSearch          bool           `toml:"deleted_taxonomies_to_search"`
		DeletedTaxonomiesToSearchPermanent bool           `toml:"deleted_taxonomies_to_search_permanent"`
	}
	Frontend struct {
		SitesPath                string   `toml:"sites_path"`
		DefaultSite              string   `toml:"default_site"`
		SecretKey                string   `toml:"secret_key"`
		CaptchaKey               string   `toml:"captcha_key"`
		CaptchaSecret            string   `toml:"captcha_secret"`
		MaxDmcaMinute            int64    `toml:"max_dmca_minute"`
		CaptchaWhiteList         []string `toml:"captcha_whitelist"`
		RouteRedirectContentItem string   `toml:"route_redirect_content_item"`
	}
	Database struct {
		Path      string `toml:"path"`
		LowMemory bool   `toml:"low_memory"`
	}
)

var apiVersionRegex = regexp.MustCompile(`^(.*)/v\d+/?$`)

func InitConfig(configPath string) {
	Config = &ConfigT{
		General: General{
			Nginx:                true,
			LangCookie:           "lng",
			Development:          runtime.GOOS == "windows",
			GeoipUrl:             "https://totaltraffictrader.com/geo/country.tar.gz",
			RecreateWorkers:      50,
			InnerRecreateWorkers: 20,
			ToplistDataUrl:       "/_toplist_data.json",
			ApiTimeout:           types.Duration(time.Second * 20),
		},
		Frontend: Frontend{
			MaxDmcaMinute: 5,
			RouteRedirectContentItem: "/_redirect_content_item",
		},
		Translations: make(map[string]map[string]string),
	}
	if _, err := toml.DecodeFile(configPath, Config); err != nil {
		log.Fatalln(configPath, ":", err)
	}
	matches := apiVersionRegex.FindStringSubmatch(Config.General.ApiUrl)
	if matches != nil {
		Config.General.ApiUrl = matches[1] + "/"
	}
	Config.MainPath = filepath.Dir(configPath)
}
