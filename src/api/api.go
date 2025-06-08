package api

import (
	"encoding/json"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
)

type Data map[string]interface{}
type Method string
type ApiUri string

var ApiHasTrouble atomic.Bool
var ApiWriteHasTrouble atomic.Bool
var apiCheckMutex sync.Mutex
var apiWriteErrCount atomic.Int32
var apiReadErrCount atomic.Int32
var errsForTrouble = int32(10)
var apiCheckRunning bool

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
	uriContentId          ApiUri = "content-id"
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
	uriCommentsGet        ApiUri = "comments"
	uriCommentsReplies    ApiUri = "comments/replies"
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
	uriHealth             ApiUri = "health"
	uriUpdateConfig       ApiUri = "update-config"
)

var ErrApiWriteTrouble = errors.New("api not available for write operations now. Try later")
var ErrApiTrouble = errors.New("api not available now. Try later")

func Request(siteDomain string, method Method, uri ApiUri, data interface{}) (response json.RawMessage, err error) {
	if ApiHasTrouble.Load() {
		//return nil, ErrApiTrouble
	}
	if method != methodGet && ApiWriteHasTrouble.Load() {
		//return nil, ErrApiWriteTrouble
	}
	if siteDomain == "" {
		siteDomain = internal.Config.Frontend.DefaultSite
	}
	siteConfigPath := filepath.Join(internal.Config.Frontend.SitesPath, siteDomain, "config.toml")
	siteConfig := internal.GetConfig(siteConfigPath, UpdateConfigRetry)
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
		if method == methodGet {
			errCount := apiReadErrCount.Add(1)
			if errCount > errsForTrouble {
				func() {
					apiCheckMutex.Lock()
					defer apiCheckMutex.Unlock()
					if apiCheckRunning {
						return
					}
					ApiHasTrouble.Store(true)
					apiCheckRunning = true
					go periodicCheckApi()
				}()
			}
		} else {
			errCount := apiWriteErrCount.Add(1)
			if errCount > errsForTrouble {
				func() {
					apiCheckMutex.Lock()
					defer apiCheckMutex.Unlock()
					if apiCheckRunning {
						return
					}
					ApiWriteHasTrouble.Store(true)
					apiCheckRunning = true
					go periodicCheckApi()
				}()
			}
		}
		return
	} else if method == methodGet {
		apiReadErrCount.Store(0)
	} else {
		apiWriteErrCount.Store(0)
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
		if !strings.Contains(errorString, "favicon.ico") && !strings.Contains(errorString, "not found") {
			log.Printf("error from api: %s, %s, %s, %s", errorString, siteDomain, method, uri)
		}
		return
	}
	response = r.Value
	return
}

func periodicCheckApi() {
	defer func() {
		apiCheckMutex.Lock()
		apiCheckRunning = false
		apiCheckMutex.Unlock()
	}()
	for {
		if ApiHasTrouble.Load() {
			if _, err := Request("", methodGet, uriHealth, nil); err == nil {
				ApiHasTrouble.Store(false)
				apiReadErrCount.Store(0)
				apiWriteErrCount.Store(0)
			}
		} else if ApiWriteHasTrouble.Load() {
			if _, err := Request("", methodPost, uriHealth, nil); err == nil {
				ApiWriteHasTrouble.Store(false)
				apiWriteErrCount.Store(0)
			}
		}
	}
}
