package api

import (
	"log"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"github.com/valyala/fasthttp"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
)

type Data map[string]interface{}
type Method string
type ApiUri string

type apiResponse struct {
	Success bool            `json:"success"`
	Value   json.RawMessage `json:"value"`
}

const (
	methodGet    Method = "GET"
	methodPost   Method = "POST"
	methodPut    Method = "PUT"
	methodDelete Method = "DELETE"
)
const (
	uriAutocomplete       ApiUri = "autocomplete"
	uriTimeframes         ApiUri = "timeframes"
	uriOptions            ApiUri = "options"
	uriContentItem        ApiUri = "content-item"
	uriTopContent         ApiUri = "top-content"
	uriCategory           ApiUri = "category"
	uriCategoryInfo       ApiUri = "category-info"
	uriChannelInfo        ApiUri = "channel-info"
	uriTopCategories      ApiUri = "top-categories"
	uriContent            ApiUri = "content"
	uriCategoriesList     ApiUri = "categories-list"
	uriModelsList         ApiUri = "models-list"
	uriModel              ApiUri = "model"
	uriChannelsList       ApiUri = "channels-list"
	uriTopSearches        ApiUri = "searches/top"
	uriRandomSearches     ApiUri = "searches/random"
	uriRelated            ApiUri = "related"
	uriDmca               ApiUri = "dmca"
	uriCountView          ApiUri = "count-view"
	uriTopCategoriesClick ApiUri = "count-click/top-categories"
	uriCategoryClick      ApiUri = "count-click/category"
	uriTopContentClick    ApiUri = "count-click/top-content"
	uriTranslate          ApiUri = "translate"
	uriLanguages          ApiUri = "languages"
	uriCountryGroups      ApiUri = "country-groups"
)

func ApiRequest(siteDomain string, method Method, uri ApiUri, data interface{}) (response json.RawMessage, err error) {
	f := helpers.Fetch(internal.Config.General.ApiUrl + "v1/" + string(uri))
	f.WithHeader(fasthttp.HeaderAuthorization, internal.Config.General.ApiSecret)
	f.WithHeader("Totaltube-Site", siteDomain)
	f.WithHeader(fasthttp.HeaderAccept, "application/json")
	f.WithTimeout(time.Duration(internal.Config.General.ApiTimeout))
	f.WithMethod(string(method))
	if method == "GET" && data != nil {
		queryParams, ok := data.(url.Values)
		if !ok {
			err = errors.New("wrong query params")
			return
		}
		f.WithQuery(queryParams)
	} else if data != nil {
		f.WithJsonData(data)
	}
	var resp []byte
	resp, err = f.Do()
	if err != nil {
		err = errors.Wrap(err, "error getting "+internal.Config.General.ApiUrl+string(uri))
		return
	}
	var r apiResponse
	err = json.Unmarshal(resp, &r)
	if err != nil {
		log.Println(err, siteDomain, method, uri, string(resp))
		return
	}
	if !r.Success {
		var errorString string
		_ = json.Unmarshal(r.Value, &errorString)
		err = errors.New("error from api: " + errorString)
		return
	}
	response = r.Value
	return
}
