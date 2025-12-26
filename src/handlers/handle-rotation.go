package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/render"
	"github.com/tidwall/gjson"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
)

func handleRotation(rotationParams rotationParams, useTrade bool, config *types.Config, r *http.Request, w http.ResponseWriter, langId string) (toReturn bool) {
	var wg sync.WaitGroup
	hostName := r.Context().Value(types.ContextKeyHostName).(string)
	if rotationParams.Type != types.CountTypeNone {
		wg.Add(1)
		// counting the click
		params := url.Values{}
		tp := config.Params.CountTypeCategory
		switch rotationParams.Type {
		case types.CountTypeCategory:
			tp = config.Params.CountTypeCategory
		case types.CountTypeTopCategories:
			tp = config.Params.CountTypeTopCategories
		case types.CountTypeTopContent:
			tp = config.Params.CountTypeTopContent
		}
		params.Add(config.Params.CountType, tp)
		if rotationParams.ContentId != 0 {
			params.Add(config.Params.ContentId, strconv.FormatInt(rotationParams.ContentId, 10))
		}
		if rotationParams.CategoryId != 0 {
			params.Add(config.Params.CategoryId, strconv.FormatInt(rotationParams.CategoryId, 10))
		}
		if rotationParams.ThumbId != -1 {
			params.Add(config.Params.CountThumbId, strconv.FormatInt(rotationParams.ThumbId, 10))
		}
		if rotationParams.Position != -1 {
			params.Add(config.Params.CountPosition, strconv.FormatInt(rotationParams.Position, 10))
		}
		countUrl := fmt.Sprintf("http://127.0.0.1:%d%s", internal.Config.General.Port, config.Routes.Out)
		countUrl += "?" + params.Encode()
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", countUrl, nil)
			if err != nil {
				return
			}
			req.Host = hostName
			req.Header.Set("Host", hostName)
			req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
			req.Header.Set(internal.Config.General.RealIpHeader, r.Header.Get(internal.Config.General.RealIpHeader))
			req.Header.Set("Referer", r.Header.Get("Referer"))
			var client = http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,  // Максимальное количество неактивных соединений, которое можно сохранять в пуле
					MaxIdleConnsPerHost: 100,  // Максимальное количество неактивных соединений с одним хостом
					MaxConnsPerHost:     2000, // Максимальное количество соединений
					DisableKeepAlives:   true,
				},
				Timeout: time.Second * 10,
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err, params, hostName)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				log.Println(resp.StatusCode, params, hostName)
				return
			}
		}()
	}
	if useTrade {
		wg.Add(1)
		go func() {
			defer wg.Done()
			referer := r.Header.Get("Referer")
			if referer == "" {
				return
			}
			parsedReferer, _ := url.Parse(referer)
			if parsedReferer == nil {
				return
			}
			parsedRefererHost := strings.TrimPrefix(strings.ToLower(parsedReferer.Host), "www.")
			if parsedRefererHost != hostName {
				log.Printf("referer %s not equal to %s", parsedRefererHost, hostName)
				return
			}
			tradeUrl := config.General.TradeUrlTemplate
			if config.General.LocalTradeUrlTemplate != "" {
				tradeUrl = config.General.LocalTradeUrlTemplate
			}
			tradeUrl = strings.ReplaceAll(tradeUrl, "{{encoded_url}}", url.QueryEscape(r.URL.Path))
			tradeUrl = strings.ReplaceAll(tradeUrl, "{{url}}", r.URL.Path)
			tradeUrl = strings.ReplaceAll(tradeUrl, "{{skim}}", rotationParams.Skim)
			tradeUrl = strings.ReplaceAll(tradeUrl, "{{lang}}", langId)
			parsedTradeUrl, err := url.Parse(tradeUrl)
			if err != nil {
				log.Println(err)
				return
			}
			if parsedTradeUrl.Host == "" {
				parsedTradeUrl.Host = hostName
				parsedTradeUrl.Scheme = "https"
				tradeUrl = parsedTradeUrl.String()
			}
			req, err := http.NewRequest("GET", tradeUrl, nil)
			if err != nil {
				log.Println(err, tradeUrl, hostName)
				return
			}
			req.Host = hostName
			req.Header = r.Header.Clone()
			req.Header.Set("Host", hostName)
			req.Header.Set(internal.Config.General.RealIpHeader, r.Header.Get(internal.Config.General.RealIpHeader))
			req.Header.Set("CF-Connecting-IP", r.Header.Get(internal.Config.General.RealIpHeader))
			req.Header.Del("Accept-Encoding")
			req.Header.Del("Connection")

			var client = http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,  // Максимальное количество неактивных соединений, которое можно сохранять в пуле
					MaxIdleConnsPerHost: 100,  // Максимальное количество неактивных соединений с одним хостом
					MaxConnsPerHost:     2000, // Максимальное количество соединений
					DisableKeepAlives:   true,
				},
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
				Timeout: time.Second * 2,
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("error doing request %s %s %s", err, tradeUrl, hostName)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == 302 {
				if resp.Header.Get("Location") != r.URL.Path {
					toReturn = true
					http.Redirect(w, r, resp.Header.Get("Location"), http.StatusTemporaryRedirect)
				}
				return
			}
			if resp.StatusCode != 200 {
				log.Printf("error doing request %s %s %s", http.StatusText(resp.StatusCode), tradeUrl, hostName)
				return
			}
			if resp.StatusCode == 200 {
				if render.GetContentType(resp.Header.Get("Content-Type")) == render.ContentTypeJSON {
					body, _ := io.ReadAll(resp.Body)
					data := gjson.ParseBytes(body)
					if data.Get("url").String() != r.URL.Path && !data.Get("filtered").Bool() {
						toReturn = true
						log.Println("redirecting to", data.Get("url").String())
						http.Redirect(w, r, data.Get("url").String(), http.StatusTemporaryRedirect)
						return
					}
					return
				}
				toReturn = true
				for k, v := range resp.Header {
					for _, vv := range v {
						w.Header().Set(k, vv)
					}
				}
				w.WriteHeader(resp.StatusCode)
				_, _ = io.Copy(w, resp.Body)
				return
			}
		}()
	}
	wg.Wait()
	return
}
