package site

type (
	Config struct {
		Routes     ConfigRoutes
		General    ConfigGeneral
		Params     ConfigParams
		Javascript ConfigJs
		Scss       ConfigScss
	}
	ConfigRoutes struct {
		TopCategories string `toml:"top_categories"`
		TopContent    string `toml:"top_content"`
		Autocomplete  string
		Search        string
		Popular       string
		New           string
		Long          string
		Model         string
		Models        string
		Category      string
		Channel       string
		Content       string
		Out           string
		Maintenance   string
		Custom        map[string]string `toml:"custom"`
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
		SortBy               string `toml:"sort_by"`
		SortByViews          string `toml:"sort_by_views"`
		SortByViewsTimeframe string `toml:"sort_by_views_timeframe"`
		SortByDuration       string `toml:"sort_by_duration"`
		SortByDate           string `toml:"sort_by_date"`
		Page                 string `toml:"page"`
		Nocache              string `toml:"nocache"`
	}
	ConfigJs struct {
		Entries []string `toml:"entries"`
	}
	ConfigScss struct {
		Entries []string `toml:"entries"`
	}
	ConfigGeneral struct {
		TradeUrlTemplate string `toml:"trade_url_template"`
		MultiLanguage    bool   `toml:"multi_language"`
		MinifyHtml       bool   `toml:"minify_html"`
		Debug            bool   `toml:"debug"`
	}
)

func NewConfig() *Config {
	var n = Config{
		Routes:  ConfigRoutes{
			TopCategories: "/",
			TopContent: "",
			Autocomplete: "/autocomplete",
			Search: "/search/:query",
			Popular: "/best",
			New:"/new",
			Long: "/long",
			Model:"/model/:slug",
			Models:"/models-list",
			Category: "/category/:slug",
			Channel: "/channel/:slug",
			Content: "/content/:category/:slug",
			Out: "/c",
			Maintenance: "/maintenance",
		},
		General: ConfigGeneral{MinifyHtml: true},
		Javascript: ConfigJs{
			Entries: []string{"main.ts"},
		},
		Scss: ConfigScss{
			Entries: []string{"main.scss"},
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
			SortBy:               "sort",
			SortByViews:          "views",
			SortByViewsTimeframe: "timeframe",
			SortByDate:           "date",
			SortByDuration:       "duration",
			Page:                 "page",
			Nocache:              "nocache",
		},
	}
	return &n
}
