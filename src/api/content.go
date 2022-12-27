package api

import (
	"net"
	"net/url"
	"strconv"

	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/types"
)

type SortBy string

const (
	SortPopular      SortBy = "popular"
	SortDated        SortBy = "dated"
	SortViews        SortBy = "views"
	SortDuration     SortBy = "duration"
	SortRand         SortBy = "rand"
	SortRandNoPaging SortBy = "rand1"
	SortTitle        SortBy = "title"
	SortTotal        SortBy = "total"
	SortNone         SortBy = ""
)

const (
	TimeframeAll   = "all"
	TimeframeMonth = "month"
	TimeframeHour  = "hour"
)

type ContentParams struct {
	Ip             net.IP
	Lang           string
	Page           int64
	Amount         int64
	CategoryId     int64
	CategorySlug   string
	ChannelId      int64
	ChannelSlug    string
	ModelId        int64
	ModelSlug      string
	RelatedMessage string // на основании которого будет выдан related контент ( для модели это title модели )
	Sort           SortBy
	Timeframe      string // таймфрейм при сортировке по views
	Tag            string
	DurationGte    int64
	DurationLt     int64
	SearchQuery    string
	IsNatural      bool   // true, если поисковый запрос создан самим пользователем, а не выбран в автокомплите
	UserAgent      string // UserAgent текущего клиента
	GroupId        int64
}

func Content(siteDomain string, params ContentParams) (results *types.ContentResults, rawResponse json.RawMessage, err error) {
	rawResponse, err = ContentRaw(siteDomain, params)
	if err != nil {
		return
	}
	results = new(types.ContentResults)
	err = json.Unmarshal(rawResponse, results)
	return
}

func ContentRaw(siteDomain string, params ContentParams) (rawResponse json.RawMessage, err error) {
	var data = url.Values{}
	if params.Ip != nil {
		data.Add("ip", params.Ip.String())
	}
	if params.Lang != "" {
		data.Add("lang", params.Lang)
	}
	if params.Page > 0 {
		data.Add("page", strconv.FormatInt(params.Page, 10))
	}
	if params.Amount > 0 {
		data.Add("amount", strconv.FormatInt(params.Amount, 10))
	}
	if params.CategoryId > 0 {
		data.Add("category_id", strconv.FormatInt(params.CategoryId, 10))
	}
	if params.CategorySlug != "" {
		data.Add("category_slug", params.CategorySlug)
	}
	if params.ChannelId > 0 {
		data.Add("channel_id", strconv.FormatInt(params.ChannelId, 10))
	}
	if params.ChannelSlug != "" {
		data.Add("channel_slug", params.ChannelSlug)
	}
	if params.ModelId > 0 {
		data.Add("model_id", strconv.FormatInt(params.ModelId, 10))
	}
	if params.ModelSlug != "" {
		data.Add("model_slug", params.ModelSlug)
	}
	if params.Sort != "" {
		data.Add("sort", string(params.Sort))
	}
	if params.Timeframe != "" {
		data.Add("timeframe", params.Timeframe)
	}
	if params.Tag != "" {
		data.Add("tag", params.Tag)
	}
	if params.DurationGte > 0 {
		data.Add("duration_gte", strconv.FormatInt(params.DurationGte, 10))
	}
	if params.DurationLt > 0 {
		data.Add("duration_lt", strconv.FormatInt(params.DurationLt, 10))
	}
	if params.SearchQuery != "" {
		data.Add("search_query", params.SearchQuery)
	}
	if params.IsNatural {
		data.Add("is_natural", strconv.FormatBool(params.IsNatural))
	}
	if params.UserAgent != "" {
		data.Add("user_agent", params.UserAgent)
	}
	if params.RelatedMessage != "" {
		data.Add("related_message", params.RelatedMessage)
	}
	rawResponse, err = ApiRequest(siteDomain, methodGet, uriContent, data)
	return
}