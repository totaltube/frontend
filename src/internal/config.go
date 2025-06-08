package internal

import (
	"log"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/samber/lo"

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
		Mail         Mail
		Comments     Comments
		Related      Related
		Translations map[string]map[string]string `toml:"translations"`
		Custom       map[string]string            `toml:"custom"`
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
		RandomizeRatio                     float64        `toml:"randomize_ratio"`
		DebugRoute                         string         `toml:"debug_route"`
		TranslateStreams                   uint16         `toml:"translate_streams"` // number of simultaneous streams for translation
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
		Path                       string `toml:"path"`
		LowMemory                  bool   `toml:"low_memory"`
		BackupPath                 string `toml:"backup_path"`
		RestoreFromBackup          bool   `toml:"restore_from_backup"`
		DebugBadger                bool   `toml:"debug_badger"`
		Engine                     string `toml:"engine"`
		NoTranslationsAccessUpdate bool   `toml:"no_translations_access_update"`
		SyncWrites                 bool   `toml:"sync_writes"`
		DetectConflicts            bool   `toml:"detect_conflicts"`
	}
	Mail struct {
		Secure       bool
		Port         int64
		Hostname     string
		User         string
		Password     string
		Timeout      uint
		AddressFrom  string `toml:"address_from"`
		AddressReply string `toml:"address_reply"`
	}
	Comments struct {
		ItemsPerPage int `toml:"items_per_page"`
		MaxReplies   int `toml:"max_replies"`
	}
	Related struct {
		TitleTranslated              *bool    `toml:"title_translated"`
		TitleTranslatedMinTermFreq   *int     `toml:"title_translated_min_term_freq"`
		TitleTranslatedMaxQueryTerms *int     `toml:"title_translated_max_query_terms"`
		TitleTranslatedBoost         *float64 `toml:"title_translated_boost"`
		Title                        *bool    `toml:"title"`
		TitleMinTermFreq             *int     `toml:"title_min_term_freq"`
		TitleMaxQueryTerms           *int     `toml:"title_max_query_terms"`
		TitleBoost                   *float64 `toml:"title_boost"`
		Tags                         *bool    `toml:"tags"`
		TagsMinTermFreq              *int     `toml:"tags_min_term_freq"`
		TagsMaxQueryTerms            *int     `toml:"tags_max_query_terms"`
		TagsBoost                    *float64 `toml:"tags_boost"`
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
			TranslateStreams:     10,
		},
		Frontend: Frontend{
			MaxDmcaMinute:            5,
			RouteRedirectContentItem: "/_redirect_content_item",
		},
		Database: Database{
			Engine:          "badger",
			SyncWrites:      true,
			DetectConflicts: true,
		},
		Translations: make(map[string]map[string]string),
		Mail: Mail{
			Secure:  false,
			Timeout: 30,
		},
		Comments: Comments{
			ItemsPerPage: 30,
			MaxReplies:   200,
		},
	}
	if _, err := toml.DecodeFile(configPath, Config); err != nil {
		log.Fatalln(configPath, ":", err)
	}
	if !lo.Contains([]string{"badger", "bolt", "pebble"}, Config.Database.Engine) {
		log.Fatalln("Unsupported database engine:", Config.Database.Engine)
	}
	if Config.General.RandomizeRatio < 0 {
		Config.General.RandomizeRatio = 0
		log.Println("Randomize ratio can't be negative, set to 0")
	}
	if Config.General.RandomizeRatio > 1 {
		Config.General.RandomizeRatio = 1
		log.Println("Randomize ratio can't be more than 1, set to 1")
	}
	matches := apiVersionRegex.FindStringSubmatch(Config.General.ApiUrl)
	if matches != nil {
		Config.General.ApiUrl = matches[1] + "/"
	}
	Config.MainPath = filepath.Dir(configPath)
	if Config.General.TranslateStreams < 1 || Config.General.TranslateStreams > 1000 {
		Config.General.TranslateStreams = 1
	}
}
