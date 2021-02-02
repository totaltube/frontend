package site

type (
	Config struct {
		Routes     ConfigRoutes
		General    ConfigGeneral
		Params     ConfigParams
		Javascript ConfigJs          `json:"-"`
		Scss       ConfigScss        `json:"-"`
		Custom     map[string]string `json:"-"`
	}
	ConfigRoutes struct {
		TopCategories    string `toml:"top_categories"`
		TopContent       string `toml:"top_content"`
		Autocomplete     string
		Search           string
		Popular          string
		New              string
		Long             string
		Model            string
		Models           string
		Category         string
		Channel          string
		ContentItem      string `toml:"content_item"`
		Out              string
		Maintenance      string
		LanguageTemplate string            `toml:"language_template"`
		Custom           map[string]string `toml:"custom"`
	}
	ConfigParams struct {
		CategorySlug         string `toml:"category_slug"`
		CategoryId           string `toml:"category_id"`
		ModelSlug            string `toml:"model_slug"`
		ModelId              string `toml:"model_id"`
		ChannelSlug          string `toml:"channel_slug"`
		ChannelId            string `toml:"channel_id"`
		DurationFrom         string `toml:"duration_from"`
		DurationTo           string `toml:"duration_to"`
		SearchQuery          string `toml:"search_query"`
		SearchNatural        string `toml:"search_natural"`
		SortBy               string `toml:"sort_by"`
		SortByViews          string `toml:"sort_by_views"`
		SortByViewsTimeframe string `toml:"sort_by_views_timeframe"`
		SortByDuration       string `toml:"sort_by_duration"`
		SortByDate           string `toml:"sort_by_date"`
		SortByRand           string `toml:"sort_by_rand"`
		Page                 string `toml:"page"`
		Nocache              string `toml:"nocache" json:"-"`
	}
	ConfigJs struct {
		Entries     []string `toml:"entries"`
		Destination string   `toml:"destination"`
		Minify      bool     `toml:"minify"`
	}
	ConfigScss struct {
		Entries     []string `toml:"entries"`
		Destination string   `toml:"destination"`
		ImagesPath  string   `toml:"images_path"`
		FontsPath   string   `toml:"fonts_path"`
		Minify      bool     `toml:"minify"`
	}
	ConfigGeneral struct {
		TradeUrlTemplate           string `toml:"trade_url_template"`
		MultiLanguage              bool   `toml:"multi_language"`
		MinifyHtml                 bool   `toml:"minify_html" json:"-"`
		PaginationMaxRenderedLinks int    `toml:"pagination_max_rendered_links"`
		Debug                      bool   `toml:"debug"`
	}
)

func NewConfig() *Config {
	var n = Config{
		Routes: ConfigRoutes{
			TopCategories:    "/",
			TopContent:       "",
			Autocomplete:     "/autocomplete",
			Search:           "/search/:query",
			Popular:          "/best",
			New:              "/new",
			Long:             "/long",
			Model:            "/model/:slug",
			Models:           "/models-list",
			Category:         "/category/:slug",
			Channel:          "/channel/:slug",
			ContentItem:      "/content/:category/:slug",
			Out:              "/c",
			Maintenance:      "/maintenance",
			LanguageTemplate: "/:lang:route",
		},
		General: ConfigGeneral{MinifyHtml: true, PaginationMaxRenderedLinks: 10},
		Javascript: ConfigJs{
			Entries: []string{"main.ts"},
			Minify:  true,
		},
		Scss: ConfigScss{
			Entries:    []string{"main.scss"},
			ImagesPath: "images",
			FontsPath:  "fonts",
			Minify:     true,
		},
		Params: ConfigParams{
			CategorySlug:         "category",
			CategoryId:           "category_id",
			ModelSlug:            "model",
			ModelId:              "model_id",
			ChannelSlug:          "channel",
			ChannelId:            "channel_id",
			DurationFrom:         "duration_from",
			DurationTo:           "duration_to",
			SearchQuery:          "q",
			SearchNatural:        "natural",
			SortBy:               "sort",
			SortByViews:          "views",
			SortByViewsTimeframe: "timeframe",
			SortByDate:           "date",
			SortByDuration:       "duration",
			SortByRand:           "rand",
			Page:                 "page",
			Nocache:              "nocache",
		},
	}
	return &n
}
