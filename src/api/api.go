package api

import (
	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/url"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"time"
)

type Data map[string]interface{}
type method string
type apiUri string

type apiResponse struct {
	Success bool            `json:"success"`
	Value   json.RawMessage `json:"value"`
}

const (
	methodGet    method = "GET"
	methodPost   method = "POST"
	methodPut    method = "PUT"
	methodDelete method = "DELETE"
)
const (
	uriAutocomplete       apiUri = "autocomplete"
	uriTimeframes         apiUri = "timeframes"
	uriOptions            apiUri = "options"
	uriContentItem        apiUri = "content-item"
	uriTopContent         apiUri = "top-content"
	uriCategory           apiUri = "category"
	uriCategoryInfo       apiUri = "category-info"
	uriChannelInfo        apiUri = "channel-info"
	uriTopCategories      apiUri = "top-categories"
	uriContent            apiUri = "content"
	uriCategoriesList     apiUri = "categories-list"
	uriModelsList         apiUri = "models-list"
	uriModel              apiUri = "model"
	uriChannelsList       apiUri = "channels-list"
	uriTopSearches        apiUri = "searches/top"
	uriRandomSearches     apiUri = "searches/random"
	uriRelated            apiUri = "related"
	uriDmca               apiUri = "dmca"
	uriCountView          apiUri = "count-view"
	uriTopCategoriesClick apiUri = "count-click/top-categories"
	uriCategoryClick      apiUri = "count-click/category"
	uriTopContentClick    apiUri = "count-click/top-content"
	uriTranslate          apiUri = "translate"
	uriLanguages          apiUri = "languages"
)

func apiRequest(siteDomain string, method method, uri apiUri, data interface{}) (response json.RawMessage, err error) {
	f := helpers.Fetch(internal.Config.General.ApiUrl + string(uri))
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
		err = errors.Wrap(err, "error getting "+internal.Config.General.ApiUrl + string(uri))
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
