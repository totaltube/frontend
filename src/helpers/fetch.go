package helpers

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"

	"sersh.com/totaltube/frontend/internal"
)

type fetchRequest struct {
	method  string
	url     string
	headers map[string]string
	query   url.Values
	data    interface{} // json post data
	timeout time.Duration
}

func newFetchRequest(u string) *fetchRequest {
	var headers = map[string]string{
		"User-Agent": "Totaltube Frontend/1.0 (+https://totaltraffictrader.com/)",
	}
	parsed, err := url.Parse(u)
	if err != nil || (parsed.Host == "" && parsed.Scheme == "") {
		u = internal.Config.General.ApiUrl + u
		headers["Authorization"] = internal.Config.General.ApiSecret
	}
	n := fetchRequest{
		method:  "GET",
		url:     u,
		headers: headers,
		query:   url.Values{},
		data:    nil,
		timeout: time.Second * 5,
	}
	return &n
}

func (f *fetchRequest) WithMethod(method string) *fetchRequest {
	f.method = strings.ToUpper(method)
	return f
}

func (f *fetchRequest) WithUrl(url string) *fetchRequest {
	f.url = url
	return f
}

func (f *fetchRequest) WithHeader(headerName, headerValue string) *fetchRequest {
	f.headers[headerName] = headerValue
	return f
}

func (f *fetchRequest) WithHeaders(headers map[string]string) *fetchRequest {
	f.headers = headers
	return f
}
func (f *fetchRequest) WithQueryParam(paramName string, paramValue string) *fetchRequest {
	f.query[paramName] = []string{paramValue}
	return f
}
func (f *fetchRequest) WithQueryString(querystring string) *fetchRequest {
	f.query, _ = url.ParseQuery(querystring)
	return f
}
func (f *fetchRequest) WithQuery(query url.Values) *fetchRequest {
	f.query = query
	return f
}

func (f *fetchRequest) WithRawData(data []byte) *fetchRequest {
	f.data = data
	return f
}

func (f *fetchRequest) WithJsonData(data interface{}) *fetchRequest {
	f.data = data
	f.headers["Content-Type"] = "application/json"
	if f.method == "GET" {
		f.method = "POST"
	}
	return f
}

func (f *fetchRequest) WithFormData(data map[string]interface{}) *fetchRequest {
	f.data = data
	f.headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"
	if f.method == "GET" {
		f.method = "POST"
	}
	return f
}

func (f *fetchRequest) WithTimeout(timeout time.Duration) *fetchRequest {
	f.timeout = timeout
	return f
}

func (f *fetchRequest) Do() (response []byte, err error) {
	client := http.Client{
		Transport: &http.Transport{DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	var body io.Reader
	if f.data != nil {
		switch d := f.data.(type) {
		case []byte:
			body = bytes.NewReader(d)
		default:
			var dd []byte
			if f.headers["Content-Type"] == "application/x-www-form-urlencoded;charset=UTF-8" {
				form := url.Values{}
				if m, ok := d.(map[string]interface{}); !ok {
					log.Printf("wrong data type - %T\n", f.data)
					return
				} else {
					for k, v := range m {
						form.Set(k, fmt.Sprintf("%v", v))
					}
				}
				dd = []byte(form.Encode())
				body = bytes.NewReader(dd)
			} else {
				dd, err = json.Marshal(f.data)
				if err != nil {
					log.Println("can't marshal to json fetch function data")
					return
				} else {
					body = bytes.NewReader(dd)
				}
			}
		}
	}
	requestUrl := f.url
	var request *http.Request
	request, err = http.NewRequestWithContext(ctx, f.method, requestUrl, body)
	if err != nil {
		log.Println("error creating client request:", err)
		return
	}
	requestQuery := request.URL.Query()
	for k, v := range f.query {
		for _, vv := range v {
			requestQuery.Add(k, vv)
		}
	}
	request.URL.RawQuery = requestQuery.Encode()
	for name, val := range f.headers {
		request.Header.Set(name, val)
	}
	var resp *http.Response
	resp, err = client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("wrong status code: %d", resp.StatusCode))
		log.Println(f.url, err)
	}
	resp.Header.Get("Accept")
	if strings.Contains(request.Header.Get("Accept"), "application/json") {
		// проверим, что возвращенный ответ также json:
		if !strings.Contains(resp.Header.Get("Content-Type"), "/json") {
			err = errors.New(fmt.Sprintf("wrong content type: %s", resp.Header.Get("Content-Type")))
			log.Println(err)
			return
		}
	}
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return
}

func (f *fetchRequest) Json() (response map[string]interface{}) {
	f.headers["Accept"] = "application/json"
	bt, err := f.Do()
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(bt, &response)
	if err != nil {
		log.Println(err, string(bt))
		return nil
	}
	return
}

func (f *fetchRequest) Raw() []byte {
	if res, err := f.Do(); err != nil {
		log.Println(err)
		return nil
	} else {
		return res
	}
}

func (f *fetchRequest) String() string {
	raw := f.Raw()
	if raw != nil {
		return string(raw)
	}
	return ""
}

func Fetch(u string) *fetchRequest {
	return newFetchRequest(u)
}
