package api

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"net/url"
	"sersh.com/totaltube/frontend/internal"
	"strings"
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
)

func apiRequest(method method, uri apiUri, data interface{}) (response json.RawMessage, err error) {
	c := &fasthttp.Client{
		MaxIdleConnDuration: time.Duration(internal.Config.General.ApiTimeout),
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Duration(internal.Config.General.ApiTimeout))
		},
		MaxConnsPerHost: 5,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)
	freq.SetRequestURI(internal.Config.General.ApiUrl + string(uri))
	freq.SetConnectionClose()
	freq.Header.SetMethod(string(method))
	freq.Header.SetCanonical([]byte(fasthttp.HeaderUserAgent), []byte("Totaltube Frontend/1.0 (+https://totaltraffictrader.com/)"))
	freq.Header.SetCanonical([]byte(fasthttp.HeaderAccept), []byte("application/json"))
	freq.Header.SetCanonical([]byte(fasthttp.HeaderAuthorization), []byte(internal.Config.General.ApiSecret))
	if method == "GET" && data != nil {
		queryParams, ok := data.(url.Values)
		if !ok {
			err = errors.New("wrong query params")
			return
		}
		freq.URI().SetQueryString(queryParams.Encode())
	} else if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			err = errors.Wrap(err, "can't marshal body data")
			log.Println(err)
			return
		}
		freq.SetBody(d)
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err = c.DoTimeout(freq, resp, time.Duration(internal.Config.General.ApiTimeout))
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode() != 200 {
		err = errors.New(fmt.Sprintf("wrong status code: %d", resp.StatusCode()))
		log.Println(err)
		return
	}
	if !strings.HasSuffix(string(resp.Header.ContentType()), "/json") {
		err = errors.New(fmt.Sprintf("wrong content type: %s", resp.Header.ContentType()))
		log.Println(err)
		return
	}
	var r apiResponse
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		log.Println(err)
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
