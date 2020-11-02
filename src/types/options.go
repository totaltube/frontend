package types

type PageLayoutT struct {
	Amount        int32 `json:"amount"`
	CastingAmount int32 `json:"casting_amount"`
	CastingStart  int32 `json:"casting_start"`
	CastingStop   int32 `json:"casting_stop"`
}
type LayoutsT struct {
	Index     PageLayoutT `json:"index"`
	Category  PageLayoutT `json:"category"`
	Search    PageLayoutT `json:"search"`
	Favorites PageLayoutT `json:"favorites"`
}
type PopularityOptions struct {
	Enable                  bool      `json:"enable"`
	MaxClicksFromIp         int32     `json:"max_clicks_from_ip"`
	CtrRound                int32     `json:"ctr_round"`
	ClickWeightCastingThumb float32   `json:"click_weight_casting_thumb"`
	CastingViews            int32     `json:"casting_views"`
	MaxHistory              int32     `json:"max_history"`
	ClickMatrix             []float32 `json:"click_matrix"`
	CastingThumbClicks      int16     `json:"casting_thumb_clicks"`
	Layouts                 LayoutsT  `json:"layouts"`
}

type Options struct {
	Popularity              PopularityOptions `json:"popularity"`
	MaxContent              int               `json:"max_content"`               // максимум контента в базе. Остальное будем удалять
	MinContentInCategory    int               `json:"min_content_in_category"`   // Не удалять контент, если он в категории, где меньше такого количества контента
	MaxPages                int               `json:"max_pages"`                 // максимум страниц для скроллинга
	RelatedAmount           int               `json:"related_amount"`            // Количество выбираемых related запросов вместе с получением контента
	RelatedMinTotal         int               `json:"searches_min_total"`        // Минимальное количество контента, соответствующего поисковому запросу для его сохранения
	RelatedMinSearches      int               `json:"related_min_searches"`      // Минимальное количество поисков по запросу для появления его в related
	AutocompleteMinTotal    int               `json:"autocomplete_min_total"`    // Минимальное количество контента, соответствующее поисковому запросу для добавления этого запроса в автокомплит
	AutocompleteMinSearches int               `json:"autocomplete_min_searches"` // Минимальное количество поисков по поисковому запросу для добавления запроса в автокомплит
}
