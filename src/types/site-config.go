package types

type (
	SiteConfig struct {
		Routes  SiteConfigRoutes
		General SiteConfigGeneral
		Params  SiteConfigParams
	}
	SiteConfigRoutes struct {
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
		Custom        map[string]string `toml:"custom"`
	}
	SiteConfigParams struct {
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
	}
	SiteConfigGeneral struct {
		TradeUrlTemplate string `toml:"trade_url_template"`
		MultiLanguage    bool   `toml:"multi_language"`
	}
)

func NewSiteConfig() *SiteConfig {
	var n = SiteConfig{
		Routes:  SiteConfigRoutes{},
		General: SiteConfigGeneral{},
		Params: SiteConfigParams{
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
		},
	}
	return &n
}
