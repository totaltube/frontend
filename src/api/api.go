package api

import (
	"encoding/json"
	"log"
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"

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
	uriBadbotRegister     ApiUri = "badbot"
	uriBadBotsList        ApiUri = "badbots"
	uriWhitelistBotsList  ApiUri = "bot-whitelist"
	uriRating             ApiUri = "rating"
)

func Request(siteDomain string, method Method, uri ApiUri, data interface{}) (response json.RawMessage, err error) {
	if siteDomain == "" {
		siteDomain = internal.Config.Frontend.DefaultSite
	}
	siteConfigPath := filepath.Join(internal.Config.Frontend.SitesPath, siteDomain, "config.toml")
	siteConfig := internal.GetConfig(siteConfigPath)
	f := helpers.SiteFetch(siteConfig)(string(uri))
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
		err = errors.New("error from api: " + errorString + ", " + string(method) + ", " + string(uri))
		log.Printf("error from api: %s, %s, %s, %s, %v", errorString, siteDomain, method, uri, data)
		return
	}
	response = r.Value
	return
}
