package helpers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/dnscache"
	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/types"

	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/internal"
)

type FetchRequest struct {
	method  string
	url     string
	config  *types.Config
	headers map[string]string
	query   url.Values
	data    interface{} // json post data
	timeout time.Duration
}

func newFetchRequest(u string, config *types.Config) *FetchRequest {
	var headers = map[string]string{
		"User-Agent": "Totaltube Frontend/1.0 (+https://totaltraffictrader.com/)",
	}
	parsed, err := url.Parse(u)
	timeout := time.Second * 5
	if err != nil || (parsed.Host == "" && parsed.Scheme == "") {
		apiUrl := internal.Config.General.ApiUrl
		apiSecret := internal.Config.General.ApiSecret
		if config.General.ApiUrl != "" {
			apiUrl = config.General.ApiUrl
		}
		if config.General.ApiSecret != "" {
			apiSecret = config.General.ApiSecret
		}
		if u == "translate" {
			timeout = time.Second * 60
		} else {
			timeout = time.Duration(internal.Config.General.ApiTimeout)
		}
		u = apiUrl + "v1/" + u
		headers["Authorization"] = apiSecret
		headers["Accept"] = "application/json"
		if config != nil {
			headers["Totaltube-Site"] = config.Hostname
		}
	}
	n := FetchRequest{
		method:  "GET",
		url:     u,
		headers: headers,
		query:   url.Values{},
		data:    nil,
		config:  config,
		timeout: timeout,
	}
	return &n
}

func (f *FetchRequest) WithMethod(method string) *FetchRequest {
	f.method = strings.ToUpper(method)
	return f
}

func (f *FetchRequest) WithUrl(url string) *FetchRequest {
	f.url = url
	return f
}

func (f *FetchRequest) WithHeader(headerName, headerValue string) *FetchRequest {
	f.headers[headerName] = headerValue
	return f
}

func (f *FetchRequest) WithHeaders(headers map[string]string) *FetchRequest {
	f.headers = headers
	return f
}
func (f *FetchRequest) WithQueryParam(paramName string, paramValue string) *FetchRequest {
	f.query[paramName] = []string{paramValue}
	return f
}
func (f *FetchRequest) WithQueryString(querystring string) *FetchRequest {
	f.query, _ = url.ParseQuery(querystring)
	return f
}
func (f *FetchRequest) WithQuery(query url.Values) *FetchRequest {
	f.query = query
	return f
}

func (f *FetchRequest) WithRawData(data []byte) *FetchRequest {
	f.data = data
	return f
}

func (f *FetchRequest) WithJsonData(data interface{}) *FetchRequest {
	f.data = data
	f.headers["Content-Type"] = "application/json"
	if f.method == "GET" {
		f.method = "POST"
	}
	return f
}

func (f *FetchRequest) WithFormData(data map[string]interface{}) *FetchRequest {
	f.data = data
	f.headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"
	if f.method == "GET" {
		f.method = "POST"
	}
	return f
}

func (f *FetchRequest) WithTimeout(seconds int64) *FetchRequest {
	f.timeout = time.Duration(seconds) * time.Second
	return f
}

var resolver = &dnscache.Resolver{}
var resolverInitialized atomic.Bool
var dnsDialer = func(ctx context.Context, network, address string) (conn net.Conn, err error) {
	if !resolverInitialized.Swap(true) {
		go func() {
			ticker := time.NewTicker(time.Minute * 5)
			defer ticker.Stop()
			for range ticker.C {
				resolver.Refresh(true)
			}
		}()
		log.Println("dns resolver initialized")
	}
	var host string
	var port string
	host, port, err = net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var ips []string
	ips, err = resolver.LookupHost(ctx, host)
	if err != nil {
		return nil, err
	}
	lo.Shuffle(ips)
	for _, ip := range ips {
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
		if err == nil {
			return
		}
	}
	return
}

func (f *FetchRequest) Do() (response []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("panic in fetch request: %v", r))
		}
	}()
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	var client = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,  // Максимальное количество неактивных соединений, которое можно сохранять в пуле
			MaxIdleConnsPerHost: 100,  // Максимальное количество неактивных соединений с одним хостом
			MaxConnsPerHost:     2000, // Максимальное количество соединений
			DisableKeepAlives:   true,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DialContext:         dnsDialer,
		},
		Timeout: f.timeout,
	}
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
					log.Println("can't marshal to json fetch function data", f.config.Hostname, err)
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
		log.Println("error creating client request:", err, f.config.Hostname)
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
		if strings.ToLower(name) == "host" {
			// for Host header we have special case
			request.Host = val
		}
		request.Header.Set(name, val)
	}
	request.Close = true
	var resp *http.Response
	resp, err = client.Do(request)
	elapsed := time.Since(started)
	if elapsed > time.Second*2 && !strings.Contains(request.URL.String(), "/translate") {
		log.Println("too long request for ", request.URL.String(), f.config.Hostname, elapsed)
	}
	if err != nil {
		log.Println(err, request.Host)
		return
	}

	if resp.StatusCode >= 300 {
		err = errors.New(fmt.Sprintf("wrong status code: %d", resp.StatusCode))
		log.Println(f.url, err, request.Host, f.config.Hostname)
	}
	if strings.Contains(request.Header.Get("Accept"), "application/json") {
		if !strings.Contains(resp.Header.Get("Content-Type"), "/json") {
			err = errors.New(fmt.Sprintf("wrong content type: %s", resp.Header.Get("Content-Type")))
			resp, _ := io.ReadAll(resp.Body)
			log.Println(err, string(resp), request.Host, f.config.Hostname)
			return
		}
	}
	response, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err, f.config.Hostname)
	}
	elapsed = time.Since(started)
	if elapsed > time.Second*2 && !strings.Contains(request.URL.String(), "/translate") {
		log.Println("too long getting response for ", request.URL.String(), request.Header.Get("Totaltube-Site"), elapsed)
	}
	return
}

func (f *FetchRequest) Json() (response map[string]interface{}) {
	f.headers["Accept"] = "application/json"
	bt, err := f.Do()
	if err != nil {
		log.Println(err, f.config.Hostname)
		return nil
	}
	err = json.Unmarshal(bt, &response)
	if err != nil {
		log.Println(err, string(bt), f.config.Hostname)
		return nil
	}
	return
}

func (f *FetchRequest) Raw() []byte {
	if res, err := f.Do(); err != nil {
		log.Println(err, f.config.Hostname)
		return nil
	} else {
		return res
	}
}

func (f *FetchRequest) String() string {
	raw := f.Raw()
	if raw != nil {
		return string(raw)
	}
	return ""
}

func SiteFetch(siteConfig *types.Config) func(u string) *FetchRequest {
	return func(u string) *FetchRequest {
		n := newFetchRequest(u, siteConfig)
		n.config = siteConfig
		return n
	}
}
