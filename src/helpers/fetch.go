package helpers

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
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
	n := fetchRequest{
		method: "GET",
		url:    u,
		headers: map[string]string{
			fasthttp.HeaderUserAgent: "Totaltube Frontend/1.0 (+https://totaltraffictrader.com/)",
		},
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
func (f *fetchRequest) WithData(data interface{}) *fetchRequest {
	f.data = data
	return f
}
func (f *fetchRequest) WithTimeout(timeout time.Duration) *fetchRequest {
	f.timeout = timeout
	return f
}

func (f *fetchRequest) Do() (response []byte, err error) {
	c := &fasthttp.Client{
		MaxIdleConnDuration: f.timeout,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, f.timeout)
		},
		MaxConnsPerHost: 10,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)
	freq.SetRequestURI(f.url)
	freq.SetConnectionClose()
	freq.Header.SetMethod(f.method)
	for name, val := range f.headers {
		freq.Header.Set(name, val)
	}
	freq.URI().SetQueryString(f.query.Encode())
	if f.data != nil {
		switch d := f.data.(type) {
		case []byte:
			freq.SetBody(d)
		default:
			var dd []byte
			dd, err = json.Marshal(f.data)
			if err != nil {
				log.Println("can't marshal to json fetch function data")
				return
			} else {
				freq.SetBody(dd)
			}
		}
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err = c.DoTimeout(freq, resp, f.timeout)
	if err != nil {
		return
	}
	if resp.StatusCode() != 200 {
		err = errors.New(fmt.Sprintf("wrong status code: %d", resp.StatusCode()))
		log.Println(f.url, err)
		return
	}
	if freq.Header.HasAcceptEncoding("application/json") {
		// проверим, что возвращенный ответ также json:
		if !strings.HasSuffix(string(resp.Header.ContentType()), "/json") {
			err = errors.New(fmt.Sprintf("wrong content type: %s", resp.Header.ContentType()))
			log.Println(err)
			return
		}
	}
	response = resp.Body()
	return
}

func (f *fetchRequest) Json() (response map[string]interface{}) {
	f.headers[fasthttp.HeaderAccept] = "application/json"
	bt, err := f.Do()
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(bt, &response)
	if err != nil {
		log.Println(err)
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
