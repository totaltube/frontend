package site

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/geoip"
	"sersh.com/totaltube/frontend/types"

	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
)

var ErrTemplateNotFound = errors.New("template not found")

type templates struct {
	sync.Mutex
	path        string
	templates   map[string]*pongo2.Template
	templateSet *pongo2.TemplateSet
	lastChange  time.Time
}

func (ts *templates) get(name string) (*pongo2.Template, error) {
	ts.Lock()
	defer ts.Unlock()
	if t, ok := ts.templates[name]; ok {
		return t, nil
	}
	// Парсим шаблон
	matches, err := filepath.Glob(filepath.Join(ts.path, "templates", "*"))
	if err != nil {
		return nil, errors.Wrap(err, "can't open "+ts.path)
	}
	for _, m := range matches {
		if strings.Split(filepath.Base(m), ".")[0] == name {
			// Found the template
			template, err := ts.templateSet.FromFile(m)
			if err != nil {
				return nil, err
			}
			return template, nil
		}
	}
	if name != "sitemap-video" {
		log.Println(name, "template not found")
	}
	return nil, ErrTemplateNotFound
}

func NewTemplates(path string) *templates {
	n := templates{path: path, templates: make(map[string]*pongo2.Template)}
	n.templateSet = pongo2.NewSet(filepath.Base(path), pongo2.DefaultLoader)
	n.templateSet.Options.LStripBlocks = true
	n.templateSet.Options.TrimBlocks = true
	host := filepath.Base(path)
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("error in templates file watching routine", r)
					}
				}()
				c := make(chan notify.EventInfo, 1)
				if err := notify.Watch(filepath.Join(path, "templates")+"/...", c, notify.Create,
					notify.Write, notify.Remove, notify.Rename); err != nil {
					log.Panicln(err)
				}
				defer notify.Stop(c)
				// waiting for signal of file changing
				info := <-c
				if internal.Config.General.Development {
					// In dev mode we invalidate all cache
					err := db.ClearCacheByPrefix("")
					if err != nil {
						log.Println(err)
					}
				} else {
					var changedTemplatePath = info.Path()
					abs, _ := filepath.Abs(filepath.Join(path, "templates"))
					if filepath.Dir(changedTemplatePath) == abs && filepath.Ext(changedTemplatePath) == ".twig" {
						templateName := strings.TrimSuffix(filepath.Base(changedTemplatePath), filepath.Ext(changedTemplatePath))
						switch templateName {
						case "top-categories", "category", "top-content", "404", "500",
							"model", "models", "content-item", "search", "fake-player", "sitemap-video":
							err := db.ClearCacheByPrefix(templateName + ":" + host + ":")
							if err != nil {
								log.Println(err)
							}
						}
						if strings.HasPrefix(templateName, "custom-") {
							err := db.ClearCacheByPrefix("custom:" + host + ":" + strings.TrimPrefix(templateName, "custom-") + ":")
							if err != nil {
								log.Println(err)
							}
						}
					}
				}
				n.Lock()
				n.lastChange = time.Now()
				n.Unlock()
				// After 1.5 seconds after last change we invalidate all template cache
				go func() {
					time.Sleep(time.Millisecond * 1500)
					n.Lock()
					defer n.Unlock()
					if !n.lastChange.After(time.Now().Add(-time.Millisecond * 1500)) {
						n.lastChange = time.Now()
						n.templates = make(map[string]*pongo2.Template)
					}
				}()
			}()
		}
	}()
	return &n
}

type siteTemplatesT struct {
	sync.Mutex
	siteTemplates map[string]*templates
}

func (st *siteTemplatesT) get(name, path string) (*pongo2.Template, error) {
	st.Lock()
	if ts, ok := st.siteTemplates[path]; ok {
		st.Unlock()
		return ts.get(name)
	}
	st.siteTemplates[path] = NewTemplates(path)
	st.Unlock()
	return st.siteTemplates[path].get(name)
}

var siteTemplates = siteTemplatesT{siteTemplates: map[string]*templates{}}

func GetTemplate(name, path string) (*pongo2.Template, error) {
	return siteTemplates.get(name, path)
}

func ParseTemplate(name, path string, config *types.Config, customContext pongo2.Context,
	nocache bool, cacheKey string, cacheTtl time.Duration,
	prepare func() (pongo2.Context, error),
	w http.ResponseWriter, r *http.Request) (parsed []byte, err error) {
	var addDynamicFunctions = func(ctx pongo2.Context) {
		ctx["set_cookie"] = func(name string, value interface{}, expire interface{}) {
			var expires = time.Now().Add(time.Minute * 60)
			if e, ok := expire.(time.Time); ok {
				expires = e
			}
			if e, ok := expire.(time.Duration); ok {
				expires = time.Now().Add(e)
			}
			if e, ok := expire.(int64); ok {
				expires = time.Now().Add(time.Hour * 24 * time.Duration(e))
			}
			if e, ok := expire.(int); ok {
				expires = time.Now().Add(time.Hour * 24 * time.Duration(e))
			}
			var cookie = &http.Cookie{
				Name:    name,
				Value:   fmt.Sprintf("%v", value),
				Expires: expires,
			}
			http.SetCookie(w, cookie)
		}
		headers := make(map[string]string)
		for k := range r.Header {
			headers[k] = r.Header.Get(k)
		}
		cookies := make(map[string]string)
		for _, cookie := range r.Cookies() {
			cookies[cookie.Name] = cookie.Value
		}
		ctx["cookies"] = cookies
		ctx["headers"] = headers
		ip := r.Context().Value("ip").(string)

		ctx["ip"] = ip
		ctx["country"] = func() string {
			country, _ := geoip.Country(net.ParseIP(ip))
			return country
		}
		ctx["country_group"] = func() types.CountryGroup {
			return internal.DetectCountryGroup(net.ParseIP(ip))
		}
		ctx["redirect_to"] = func(params ...interface{}) {
			var url string
			var code = http.StatusFound
			if len(params) == 0 {
				return
			}
			if len(params) >= 1 {
				url = fmt.Sprintf("%v", params[0])
			}
			if len(params) >= 2 {
				code1, _ := strconv.ParseInt(fmt.Sprintf("%v", params[1]), 10, 32)
				code = int(code1)
				if code < 300 || code > 399 {
					code = http.StatusFound
				}
			}
			http.Redirect(w, r, url, code)
			if internal.Config.General.EnableAccessLog {
				log.Println("Redirected to", url)
			}
		}
		ctx["custom_send"] = func(params ...any) {
			if len(params) == 0 {
				return
			}
			var data string
			var code = http.StatusOK
			var headers = make(map[string]string)
			if len(params) >= 1 {
				data = fmt.Sprintf("%v", params[0])
			}
			k := ""
			v := ""
			for i := 1; i < len(params); i += 2 {
				if i+1 < len(params) {
					k = fmt.Sprintf("%v", params[i])
					v = fmt.Sprintf("%v", params[i+1])
					if k == "status" {
						code1, _ := strconv.ParseInt(v, 10, 32)
						code = int(code1)
						continue
					}
					headers[k] = v
				}
			}
			if len(params) >= 2 {
				code1, _ := strconv.ParseInt(fmt.Sprintf("%v", params[1]), 10, 32)
				code = int(code1)
			}
			for k := range headers {
				w.Header().Add(k, headers[k])
			}
			w.WriteHeader(code)
			_, _ = w.Write([]byte(data))
		}
	}
	// Adding custom functions to context
	var addCustomFunctions = func(c pongo2.Context) {
		matches, _ := filepath.Glob(filepath.Join(path, "extensions/function-*.js"))
		for _, m := range matches {
			baseName := filepath.Base(m)
			funcName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(baseName, "function-"), ".js"))
			if funcName == "" {
				continue
			}
			var source = getJsSource(m)
			if len(source) == 0 {
				continue
			}
			c[funcName] = func(args ...interface{}) interface{} {
				//helpers.KeyMutex.Lock(baseName)
				//defer helpers.KeyMutex.Unlock(baseName)
				vm := gojaVmPool.Get().(*goja.Runtime)
				defer func() {
					_ = vm.Set("config", goja.Undefined())
					_ = vm.Set("fetch", goja.Undefined())
					_ = vm.Set("nocache", goja.Undefined())
					for k := range c {
						if !lo.Contains([]string{"cache", "URL", "faker"}, k) {
							_ = vm.Set(k, goja.Undefined())
						}
					}
					gojaVmPool.Put(vm)
				}()
				_ = vm.Set("config", config)
				_ = vm.Set("fetch", helpers.SiteFetch(config))
				_ = vm.Set("nocache", nocache)
				for k, v := range c {
					_ = vm.Set(k, v)
				}
				var argsString string
				var argsNameArray = make([]string, 0, len(args))
				for argIndex, arg := range args {
					var argName = fmt.Sprintf("arg%d", argIndex)
					_ = vm.Set(argName, arg)
					argsNameArray = append(argsNameArray, argName)
				}
				argsString = strings.Join(argsNameArray, ",")
				var program *goja.Program
				var err error
				program, err = getJsProgram("function:"+funcName, string(source)+" "+funcName+"("+argsString+")")
				if err != nil {
					log.Println(err)
					return nil
				}
				var v goja.Value
				v, err = vm.RunProgram(program)
				if err != nil {
					log.Println(err, path, name, config.Hostname)
					return nil
				}
				return v.Export()
			}
		}
	}
	var extendedTtl = time.Duration(math.Max(float64(time.Minute*5), float64(cacheTtl)))
	var dataCtx pongo2.Context
	dataCtx, err = prepare()
	if err != nil {
		if !strings.Contains(err.Error(), "redirect to") && !strings.Contains(err.Error(), "not found") {
			log.Println(err, path, name, config.Hostname)
		}
		return
	}
	customContext.Update(dataCtx)
	var cached []byte
	// copy custom context for GetCachedTimeout recreate function call to avoid concurrent map write
	var customContextCopy = make(pongo2.Context)
	for k, v := range customContext {
		customContextCopy[k] = v
	}
	recreateFunc := func() (result []byte, err error) {
		c := generateContext(name, path, customContextCopy)
		addCustomFunctions(c)
		var template *pongo2.Template
		template, err = GetTemplate(name, path)
		if err != nil {
			if err != ErrTemplateNotFound {
				log.Println(err, name, path, config.Hostname)
			}
			return
		}
		result, err = template.ExecuteBytes(c)
		if err != nil {
			log.Println(err, name, path, config.Hostname)
			return
		}
		// выведем строку из result, которая содержит слово dynamic
		if config.General.MinifyHtml {
			result = helpers.MinifyBytes(result)
		}
		return
	}
	if cacheTtl > 0 {
		cached, err = db.GetCachedTimeout(cacheKey, cacheTtl, extendedTtl, recreateFunc, nocache)
	} else {
		cached, err = recreateFunc()
	}
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Println(err, path, name, config.Hostname)
		}
		return
	}
	c := generateContext(name, path, customContext)
	addCustomFunctions(c)
	addDynamicFunctions(c)
	parsed = InsertDynamic(cached, path, c)
	parsed = postHook(parsed, name, path, config, c, nocache)
	return
}
