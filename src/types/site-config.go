package types

type (
	Config struct {
		Routes     ConfigRoutes
		General    ConfigGeneral
		Params     ConfigParams
		Javascript ConfigJs          `json:"-"`
		Scss       ConfigScss        `json:"-"`
		Custom     map[string]string `json:"-"`
		Hostname   string            `json:"-"`
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
		FakePlayer       string `toml:"fake_player"`
		Out              string
		Dmca             string
		VideoEmbed       string            `toml:"video_embed"`
		LanguageTemplate string            `toml:"language_template"`
		Custom           map[string]string `toml:"custom"`
	}
	ConfigParams struct {
		ContentSlug            string `toml:"content_slug"`
		ContentId              string `toml:"content_id"`
		CountView              string `toml:"count_view"`
		CategorySlug           string `toml:"category_slug"`
		CategoryId             string `toml:"category_id"`
		ModelSlug              string `toml:"model_slug"`
		ModelId                string `toml:"model_id"`
		ChannelSlug            string `toml:"channel_slug"`
		ChannelId              string `toml:"channel_id"`
		DurationGte            string `toml:"duration_gte"`
		DurationLt             string `toml:"duration_lt"`
		SearchQuery            string `toml:"search_query"`
		SearchNatural          string `toml:"search_natural"`
		SortBy                 string `toml:"sort_by"`
		SortByViews            string `toml:"sort_by_views"`
		SortByViewsTimeframe   string `toml:"sort_by_views_timeframe"`
		SortByDuration         string `toml:"sort_by_duration"`
		SortByDate             string `toml:"sort_by_date"`
		SortByRand             string `toml:"sort_by_rand"`
		Page                   string `toml:"page"`
		CountType              string `toml:"count_type"`
		CountRedirect          string `toml:"count_redirect"`
		CountTypeCategory      string `toml:"count_type_category"`
		CountTypeTopCategories string `toml:"count_type_top_categories"`
		CountTypeTopContent    string `toml:"count_type_top_content"`
		CountTypeCategoryView  string `toml:"count_type_category_view"`
		CountThumbId           string `toml:"count_thumb_id"`
		Nocache                string `toml:"nocache" json:"-"`
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
		CanonicalUrl               string `toml:"canonical_url"`
		TradeUrlTemplate           string `toml:"trade_url_template"`
		ModelsPerPage              int    `toml:"models_per_page"`
		ContentRelatedAmount       int    `toml:"content_related_amount"`
		FakeVideoPage              bool   `toml:"fake_video_page"`
		MultiLanguage              bool   `toml:"multi_language"`
		DefaultLanguage            string `toml:"default_language"`
		NoRedirectDefaultLanguage  bool   `toml:"no_redirect_default_language"`
		MinifyHtml                 bool   `toml:"minify_html" json:"-"`
		PaginationMaxRenderedLinks int    `toml:"pagination_max_rendered_links"`
		DisableCategoriesRedirect  bool   `toml:"disable_categories_redirect"`
		Debug                      bool   `toml:"debug"`
		ApiUrl                     string `toml:"api_url"`
		ApiSecret                  string `toml:"api_secret"`
	}
)

func NewConfig() *Config {
	var n = Config{
		Routes: ConfigRoutes{
			TopCategories:    "/",
			TopContent:       "",
			Autocomplete:     "/autocomplete",
			Search:           "/search/{query}",
			Popular:          "/best",
			New:              "/new",
			Long:             "/long",
			Model:            "/model/{slug}",
			Models:           "/models-list",
			Category:         "/category/{slug}",
			Channel:          "/channel/{slug}",
			ContentItem:      "/content/{category}/{slug}",
			FakePlayer:       "/player/{data}/{hash}",
			Dmca:             "/dmca",
			Out:              "/c",
			LanguageTemplate: "/{lang}{route}",
		},
		General: ConfigGeneral{
			MinifyHtml:                 true,
			PaginationMaxRenderedLinks: 10,
			ModelsPerPage:              200,
			ContentRelatedAmount:       16,
			DefaultLanguage:            "en",
		},
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
			ContentId:              "id",
			ContentSlug:            "slug",
			CategorySlug:           "category",
			CategoryId:             "cid",
			ModelSlug:              "model",
			ModelId:                "model_id",
			ChannelSlug:            "channel",
			ChannelId:              "channel_id",
			DurationGte:            "duration_from",
			DurationLt:             "duration_to",
			SearchQuery:            "q",
			SearchNatural:          "natural",
			SortBy:                 "sort",
			SortByViews:            "views",
			SortByViewsTimeframe:   "timeframe",
			SortByDate:             "date",
			SortByDuration:         "duration",
			SortByRand:             "rand",
			Page:                   "page",
			CountRedirect:          "r",
			CountType:              "t",
			CountTypeCategory:      "c",
			CountTypeTopContent:    "tc",
			CountTypeTopCategories: "tca",
			CountTypeCategoryView:  "ccv",
			CountThumbId:           "tid",
			Nocache:                "nocache",
			CountView:              "cv",
		},
	}
	return &n
}
