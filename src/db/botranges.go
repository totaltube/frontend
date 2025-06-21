package db

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"github.com/tidwall/gjson"
	"sersh.com/totaltube/frontend/api"
	"sersh.com/totaltube/frontend/helpers"
)

// GetSEBotRanges gets SE bot ranges
func GetSEBotRanges() (ranges []string, err error) {
	var resultRaw []byte
	resultRaw, err = GetCachedTimeout("se_bot_ranges", time.Hour*24, time.Hour*100500, func() (result []byte, err error) {
		var ipRanges = make([]string, 0, 10000)
		for _, u := range []string{"https://developers.google.com/search/apis/ipranges/googlebot.json", "https://developers.google.com/static/search/apis/ipranges/special-crawlers.json", "https://www.bing.com/toolbox/bingbot.json"} {
			var req *http.Request
			req, _ = http.NewRequest("GET", u, nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0")
			req.Close = true
			var resp *http.Response
			client := &http.Client{
				Timeout: time.Second * 60,
			}
			resp, err = client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}

			var bt []byte
			bt, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return
			}
			data := gjson.ParseBytes(bt)
			data.Get("prefixes").ForEach(func(_, value gjson.Result) bool {
				ipv4Prefix := value.Get("ipv4Prefix").String()
				ipv6Prefix := value.Get("ipv6Prefix").String()
				if ipv4Prefix != "" {
					ipRanges = append(ipRanges, ipv4Prefix)
				}
				if ipv6Prefix != "" {
					ipRanges = append(ipRanges, ipv6Prefix)
				}
				return true
			})
		}
		// parsing also ip ranges for duckduckgo
		data := gjson.Parse(`{"creationTime":"2023-08-08T22:03:18.000000","prefixes":[{"ipv6Prefix":"2001:4860:4801:10::/64"},{"ipv6Prefix":"2001:4860:4801:11::/64"},{"ipv6Prefix":"2001:4860:4801:12::/64"},{"ipv6Prefix":"2001:4860:4801:13::/64"},{"ipv6Prefix":"2001:4860:4801:14::/64"},{"ipv6Prefix":"2001:4860:4801:15::/64"},{"ipv6Prefix":"2001:4860:4801:16::/64"},{"ipv6Prefix":"2001:4860:4801:17::/64"},{"ipv6Prefix":"2001:4860:4801:18::/64"},{"ipv6Prefix":"2001:4860:4801:19::/64"},{"ipv6Prefix":"2001:4860:4801:1a::/64"},{"ipv6Prefix":"2001:4860:4801:1b::/64"},{"ipv6Prefix":"2001:4860:4801:1c::/64"},{"ipv6Prefix":"2001:4860:4801:1d::/64"},{"ipv6Prefix":"2001:4860:4801:20::/64"},{"ipv6Prefix":"2001:4860:4801:21::/64"},{"ipv6Prefix":"2001:4860:4801:22::/64"},{"ipv6Prefix":"2001:4860:4801:23::/64"},{"ipv6Prefix":"2001:4860:4801:24::/64"},{"ipv6Prefix":"2001:4860:4801:25::/64"},{"ipv6Prefix":"2001:4860:4801:26::/64"},{"ipv6Prefix":"2001:4860:4801:27::/64"},{"ipv6Prefix":"2001:4860:4801:28::/64"},{"ipv6Prefix":"2001:4860:4801:29::/64"},{"ipv6Prefix":"2001:4860:4801:2::/64"},{"ipv6Prefix":"2001:4860:4801:2a::/64"},{"ipv6Prefix":"2001:4860:4801:2b::/64"},{"ipv6Prefix":"2001:4860:4801:2c::/64"},{"ipv6Prefix":"2001:4860:4801:2d::/64"},{"ipv6Prefix":"2001:4860:4801:2e::/64"},{"ipv6Prefix":"2001:4860:4801:2f::/64"},{"ipv6Prefix":"2001:4860:4801:30::/64"},{"ipv6Prefix":"2001:4860:4801:31::/64"},{"ipv6Prefix":"2001:4860:4801:32::/64"},{"ipv6Prefix":"2001:4860:4801:33::/64"},{"ipv6Prefix":"2001:4860:4801:34::/64"},{"ipv6Prefix":"2001:4860:4801:35::/64"},{"ipv6Prefix":"2001:4860:4801:36::/64"},{"ipv6Prefix":"2001:4860:4801:37::/64"},{"ipv6Prefix":"2001:4860:4801:38::/64"},{"ipv6Prefix":"2001:4860:4801:39::/64"},{"ipv6Prefix":"2001:4860:4801:3::/64"},{"ipv6Prefix":"2001:4860:4801:3a::/64"},{"ipv6Prefix":"2001:4860:4801:3b::/64"},{"ipv6Prefix":"2001:4860:4801:3c::/64"},{"ipv6Prefix":"2001:4860:4801:3d::/64"},{"ipv6Prefix":"2001:4860:4801:3e::/64"},{"ipv6Prefix":"2001:4860:4801:40::/64"},{"ipv6Prefix":"2001:4860:4801:41::/64"},{"ipv6Prefix":"2001:4860:4801:42::/64"},{"ipv6Prefix":"2001:4860:4801:43::/64"},{"ipv6Prefix":"2001:4860:4801:44::/64"},{"ipv6Prefix":"2001:4860:4801:45::/64"},{"ipv6Prefix":"2001:4860:4801:46::/64"},{"ipv6Prefix":"2001:4860:4801:47::/64"},{"ipv6Prefix":"2001:4860:4801:48::/64"},{"ipv6Prefix":"2001:4860:4801:49::/64"},{"ipv6Prefix":"2001:4860:4801:4a::/64"},{"ipv6Prefix":"2001:4860:4801:50::/64"},{"ipv6Prefix":"2001:4860:4801:51::/64"},{"ipv6Prefix":"2001:4860:4801:53::/64"},{"ipv6Prefix":"2001:4860:4801:54::/64"},{"ipv6Prefix":"2001:4860:4801:55::/64"},{"ipv6Prefix":"2001:4860:4801:60::/64"},{"ipv6Prefix":"2001:4860:4801:61::/64"},{"ipv6Prefix":"2001:4860:4801:62::/64"},{"ipv6Prefix":"2001:4860:4801:63::/64"},{"ipv6Prefix":"2001:4860:4801:64::/64"},{"ipv6Prefix":"2001:4860:4801:65::/64"},{"ipv6Prefix":"2001:4860:4801:66::/64"},{"ipv6Prefix":"2001:4860:4801:67::/64"},{"ipv6Prefix":"2001:4860:4801:68::/64"},{"ipv6Prefix":"2001:4860:4801:69::/64"},{"ipv6Prefix":"2001:4860:4801:6a::/64"},{"ipv6Prefix":"2001:4860:4801:6b::/64"},{"ipv6Prefix":"2001:4860:4801:6c::/64"},{"ipv6Prefix":"2001:4860:4801:6d::/64"},{"ipv6Prefix":"2001:4860:4801:6e::/64"},{"ipv6Prefix":"2001:4860:4801:6f::/64"},{"ipv6Prefix":"2001:4860:4801:70::/64"},{"ipv6Prefix":"2001:4860:4801:71::/64"},{"ipv6Prefix":"2001:4860:4801:72::/64"},{"ipv6Prefix":"2001:4860:4801:73::/64"},{"ipv6Prefix":"2001:4860:4801:74::/64"},{"ipv6Prefix":"2001:4860:4801:75::/64"},{"ipv6Prefix":"2001:4860:4801:76::/64"},{"ipv6Prefix":"2001:4860:4801:77::/64"},{"ipv6Prefix":"2001:4860:4801:78::/64"},{"ipv6Prefix":"2001:4860:4801:80::/64"},{"ipv6Prefix":"2001:4860:4801:81::/64"},{"ipv6Prefix":"2001:4860:4801:82::/64"},{"ipv6Prefix":"2001:4860:4801:83::/64"},{"ipv6Prefix":"2001:4860:4801:84::/64"},{"ipv6Prefix":"2001:4860:4801:85::/64"},{"ipv6Prefix":"2001:4860:4801:86::/64"},{"ipv6Prefix":"2001:4860:4801:87::/64"},{"ipv6Prefix":"2001:4860:4801:88::/64"},{"ipv6Prefix":"2001:4860:4801:90::/64"},{"ipv6Prefix":"2001:4860:4801:91::/64"},{"ipv6Prefix":"2001:4860:4801:92::/64"},{"ipv6Prefix":"2001:4860:4801:93::/64"},{"ipv6Prefix":"2001:4860:4801::/64"},{"ipv6Prefix":"2001:4860:4801:c::/64"},{"ipv6Prefix":"2001:4860:4801:f::/64"},{"ipv4Prefix":"192.178.5.0/27"},{"ipv4Prefix":"34.100.182.96/28"},{"ipv4Prefix":"34.101.50.144/28"},{"ipv4Prefix":"34.118.254.0/28"},{"ipv4Prefix":"34.118.66.0/28"},{"ipv4Prefix":"34.126.178.96/28"},{"ipv4Prefix":"34.146.150.144/28"},{"ipv4Prefix":"34.147.110.144/28"},{"ipv4Prefix":"34.151.74.144/28"},{"ipv4Prefix":"34.152.50.64/28"},{"ipv4Prefix":"34.154.114.144/28"},{"ipv4Prefix":"34.155.98.32/28"},{"ipv4Prefix":"34.165.18.176/28"},{"ipv4Prefix":"34.175.160.64/28"},{"ipv4Prefix":"34.176.130.16/28"},{"ipv4Prefix":"34.22.85.0/27"},{"ipv4Prefix":"34.64.82.64/28"},{"ipv4Prefix":"34.65.242.112/28"},{"ipv4Prefix":"34.80.50.80/28"},{"ipv4Prefix":"34.88.194.0/28"},{"ipv4Prefix":"34.89.10.80/28"},{"ipv4Prefix":"34.89.198.80/28"},{"ipv4Prefix":"34.96.162.48/28"},{"ipv4Prefix":"35.247.243.240/28"},{"ipv4Prefix":"66.249.64.0/27"},{"ipv4Prefix":"66.249.64.128/27"},{"ipv4Prefix":"66.249.64.160/27"},{"ipv4Prefix":"66.249.64.192/27"},{"ipv4Prefix":"66.249.64.224/27"},{"ipv4Prefix":"66.249.64.32/27"},{"ipv4Prefix":"66.249.64.64/27"},{"ipv4Prefix":"66.249.64.96/27"},{"ipv4Prefix":"66.249.65.0/27"},{"ipv4Prefix":"66.249.65.128/27"},{"ipv4Prefix":"66.249.65.160/27"},{"ipv4Prefix":"66.249.65.192/27"},{"ipv4Prefix":"66.249.65.224/27"},{"ipv4Prefix":"66.249.65.32/27"},{"ipv4Prefix":"66.249.65.64/27"},{"ipv4Prefix":"66.249.65.96/27"},{"ipv4Prefix":"66.249.66.0/27"},{"ipv4Prefix":"66.249.66.128/27"},{"ipv4Prefix":"66.249.66.160/27"},{"ipv4Prefix":"66.249.66.192/27"},{"ipv4Prefix":"66.249.66.32/27"},{"ipv4Prefix":"66.249.66.64/27"},{"ipv4Prefix":"66.249.66.96/27"},{"ipv4Prefix":"66.249.68.0/27"},{"ipv4Prefix":"66.249.68.32/27"},{"ipv4Prefix":"66.249.68.64/27"},{"ipv4Prefix":"66.249.69.0/27"},{"ipv4Prefix":"66.249.69.128/27"},{"ipv4Prefix":"66.249.69.160/27"},{"ipv4Prefix":"66.249.69.192/27"},{"ipv4Prefix":"66.249.69.224/27"},{"ipv4Prefix":"66.249.69.32/27"},{"ipv4Prefix":"66.249.69.64/27"},{"ipv4Prefix":"66.249.69.96/27"},{"ipv4Prefix":"66.249.70.0/27"},{"ipv4Prefix":"66.249.70.128/27"},{"ipv4Prefix":"66.249.70.160/27"},{"ipv4Prefix":"66.249.70.192/27"},{"ipv4Prefix":"66.249.70.224/27"},{"ipv4Prefix":"66.249.70.32/27"},{"ipv4Prefix":"66.249.70.64/27"},{"ipv4Prefix":"66.249.70.96/27"},{"ipv4Prefix":"66.249.71.0/27"},{"ipv4Prefix":"66.249.71.128/27"},{"ipv4Prefix":"66.249.71.160/27"},{"ipv4Prefix":"66.249.71.192/27"},{"ipv4Prefix":"66.249.71.224/27"},{"ipv4Prefix":"66.249.71.32/27"},{"ipv4Prefix":"66.249.71.64/27"},{"ipv4Prefix":"66.249.71.96/27"},{"ipv4Prefix":"66.249.72.0/27"},{"ipv4Prefix":"66.249.72.128/27"},{"ipv4Prefix":"66.249.72.160/27"},{"ipv4Prefix":"66.249.72.192/27"},{"ipv4Prefix":"66.249.72.224/27"},{"ipv4Prefix":"66.249.72.32/27"},{"ipv4Prefix":"66.249.72.64/27"},{"ipv4Prefix":"66.249.72.96/27"},{"ipv4Prefix":"66.249.73.0/27"},{"ipv4Prefix":"66.249.73.128/27"},{"ipv4Prefix":"66.249.73.160/27"},{"ipv4Prefix":"66.249.73.192/27"},{"ipv4Prefix":"66.249.73.224/27"},{"ipv4Prefix":"66.249.73.32/27"},{"ipv4Prefix":"66.249.73.64/27"},{"ipv4Prefix":"66.249.73.96/27"},{"ipv4Prefix":"66.249.74.0/27"},{"ipv4Prefix":"66.249.74.128/27"},{"ipv4Prefix":"66.249.74.32/27"},{"ipv4Prefix":"66.249.74.64/27"},{"ipv4Prefix":"66.249.74.96/27"},{"ipv4Prefix":"66.249.75.0/27"},{"ipv4Prefix":"66.249.75.128/27"},{"ipv4Prefix":"66.249.75.160/27"},{"ipv4Prefix":"66.249.75.192/27"},{"ipv4Prefix":"66.249.75.224/27"},{"ipv4Prefix":"66.249.75.32/27"},{"ipv4Prefix":"66.249.75.64/27"},{"ipv4Prefix":"66.249.75.96/27"},{"ipv4Prefix":"66.249.76.0/27"},{"ipv4Prefix":"66.249.76.128/27"},{"ipv4Prefix":"66.249.76.160/27"},{"ipv4Prefix":"66.249.76.192/27"},{"ipv4Prefix":"66.249.76.224/27"},{"ipv4Prefix":"66.249.76.32/27"},{"ipv4Prefix":"66.249.76.64/27"},{"ipv4Prefix":"66.249.76.96/27"},{"ipv4Prefix":"66.249.77.0/27"},{"ipv4Prefix":"66.249.77.128/27"},{"ipv4Prefix":"66.249.77.160/27"},{"ipv4Prefix":"66.249.77.192/27"},{"ipv4Prefix":"66.249.77.32/27"},{"ipv4Prefix":"66.249.77.64/27"},{"ipv4Prefix":"66.249.77.96/27"},{"ipv4Prefix":"66.249.78.0/27"},{"ipv4Prefix":"66.249.79.0/27"},{"ipv4Prefix":"66.249.79.128/27"},{"ipv4Prefix":"66.249.79.160/27"},{"ipv4Prefix":"66.249.79.192/27"},{"ipv4Prefix":"66.249.79.224/27"},{"ipv4Prefix":"66.249.79.32/27"},{"ipv4Prefix":"66.249.79.64/27"},{"ipv4Prefix":"66.249.79.96/27"}]}`)
		data.Get("prefixes").ForEach(func(_, value gjson.Result) bool {
			ipv4Prefix := value.Get("ipv4Prefix").String()
			ipv6Prefix := value.Get("ipv6Prefix").String()
			if ipv4Prefix != "" {
				ipRanges = append(ipRanges, ipv4Prefix)
			}
			if ipv6Prefix != "" {
				ipRanges = append(ipRanges, ipv6Prefix)
			}
			return true
		})
		var whitelistData []string
		whitelistData, err = api.GetWhitelistBots()
		if err != nil {
			log.Println(err)
			return
		}
		ipRanges = append(ipRanges, whitelistData...)
		result = helpers.ToJSON(ipRanges)
		return
	}, false)
	if err != nil {
		log.Println(err)
		return
	}
	arr := gjson.ParseBytes(resultRaw).Array()
	ranges = make([]string, 0, len(arr))
	for _, el := range arr {
		ranges = append(ranges, el.String())
	}
	return
}

var seTrieIpv4 *ipaddr.Trie[*ipaddr.Address]
var seTrieIpv6 *ipaddr.Trie[*ipaddr.Address]
var seCheckMutex sync.Mutex
var lastSeTrieUpdate = time.Now().Add(-time.Hour * 24)

// CheckIfSeBot checks if the IP is a SE bot
func CheckIfSeBot(ip string) (isSEBot bool, err error) {
	seCheckMutex.Lock()
	var initialized = seTrieIpv4 != nil
	if lastSeTrieUpdate.Before(time.Now().Add(-time.Hour)) || seTrieIpv4 == nil {
		lastSeTrieUpdate = time.Now()
		if initialized {
			seCheckMutex.Unlock()
		}
		// updating seTrieIpv4 structure
		var ranges []string
		ranges, err = GetSEBotRanges()
		if err != nil {
			log.Println(err)
			if !initialized {
				seCheckMutex.Unlock()
			}
			return
		}
		if initialized {
			seCheckMutex.Lock()
		}
		seTrieIpv4 = ipaddr.NewTrie[*ipaddr.Address]()
		seTrieIpv6 = ipaddr.NewTrie[*ipaddr.Address]()
		for _, r := range ranges {
			addr := ipaddr.NewIPAddressString(r).GetAddress().ToAddressBase()
			if addr.IsIPv4() {
				seTrieIpv4.Add(addr)
				continue
			}
			seTrieIpv6.Add(addr)
		}
		if seTrieIpv4.IsEmpty() {
			seTrieIpv4.Add(ipaddr.NewIPAddressString("8.8.8.8").GetAddress().ToAddressBase())
		}
		if seTrieIpv6.IsEmpty() {
			seTrieIpv6.Add(ipaddr.NewIPAddressString("::2").GetAddress().ToAddressBase())
		}
	}
	defer seCheckMutex.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			log.Println(err)
			return
		}
	}()
	addr := ipaddr.NewIPAddressString(ip).GetAddress().ToAddressBase()
	var triePath *ipaddr.ContainmentPath[*ipaddr.Address]
	if addr.IsIPv4() {
		triePath = seTrieIpv4.ElementsContaining(addr)
	} else {
		triePath = seTrieIpv6.ElementsContaining(addr)
	}
	if triePath.Count() > 0 {
		isSEBot = true
	} else {
		isSEBot = false
	}
	return
}

var botTrieIpv4 *ipaddr.Trie[*ipaddr.Address]
var botTrieIpv6 *ipaddr.Trie[*ipaddr.Address]
var botCheckMutex sync.Mutex
var lastBotTrieUpdate = time.Now().Add(-time.Hour * 24)

// CheckIfBadBot checks if the IP is a bad bot
func CheckIfBadBot(ip string, userAgent string) (isBadBot bool, err error) {
	botCheckMutex.Lock()
	if lastBotTrieUpdate.Before(time.Now().Add(-time.Minute*5)) || botTrieIpv4 == nil {
		lastBotTrieUpdate = time.Now()
		var initialized = botTrieIpv4 != nil
		if initialized {
			// если уже инициализированы боты какими-либо значениями, то используем их временно в других горутинах,
			// пока обновляем
			botCheckMutex.Unlock()
		}
		cacheKey := fmt.Sprintf("bad_bots_list:%s:%s", ip, userAgent)
		var cached []byte
		cached, err = GetCachedTimeout(cacheKey, time.Minute*10, time.Hour*100500, func() (result []byte, err error) {
			var bots []string
			bots, err = api.GetBadBots()
			if err != nil {
				log.Println(err)
				return
			}
			result = helpers.ToJSON(bots)
			return
		}, false)
		if err != nil {
			log.Println(err)
			if !initialized {
				botCheckMutex.Unlock()
			}
			return
		}
		var bots []string
		helpers.FromJSON(cached, &bots)
		if initialized {
			botCheckMutex.Lock()
		}
		botTrieIpv4 = ipaddr.NewTrie[*ipaddr.Address]()
		botTrieIpv6 = ipaddr.NewTrie[*ipaddr.Address]()
		for _, r := range bots {
			addr := ipaddr.NewIPAddressString(r).GetAddress().ToAddressBase()
			if addr.IsIPv4() {
				botTrieIpv4.Add(addr)
			} else {
				botTrieIpv6.Add(addr)
			}
		}
		if botTrieIpv4.IsEmpty() {
			botTrieIpv4.Add(ipaddr.NewIPAddressString("127.0.0.3").GetAddress().ToAddressBase())
		}
		if botTrieIpv6.IsEmpty() {
			botTrieIpv6.Add(ipaddr.NewIPAddressString("::3").GetAddress().ToAddressBase())
		}
	}
	addr := ipaddr.NewIPAddressString(ip).GetAddress().ToAddressBase()
	var triePath *ipaddr.ContainmentPath[*ipaddr.Address]
	if addr.IsIPv4() {
		triePath = botTrieIpv4.ElementsContaining(addr)
	} else {
		triePath = botTrieIpv6.ElementsContaining(addr)
	}
	if triePath.Count() > 0 {
		isBadBot = true
	}
	botCheckMutex.Unlock()
	return
}
