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
	"time"

	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/geoip"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"

	"github.com/dop251/goja"
	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
)

type redirectRet struct {
	url  string
	code int
}

type customSendRet struct {
	data    []byte
	headers http.Header
	code    int
}

func doRedirect(url string, code ...int) (r redirectRet) {
	r.url = url
	r.code = 302
	if len(code) > 0 && code[0] == 301 {
		r.code = code[0]
	}
	return
}

type ErrSendResponse struct {
	Redirect string
	JSON     interface{}
	Text     string
	Code     int
	Data     []byte
	Headers  http.Header
}

func (e ErrSendResponse) Error() string {
	return "custom response"
}

func ParseCustomTemplate(name, path string, config *types.Config,
	customContext pongo2.Context, nocache bool, w http.ResponseWriter, r *http.Request) (parsed []byte, err error) {
	extensionFile := filepath.Join(path, "extensions/route-"+name+".js")
	var source = getJsSource(extensionFile)
	if len(source) == 0 {
		err = errors.New("source template " + extensionFile + " is empty or not exists")
		return
	}

	vm := gojaVmPool.Get().(*goja.Runtime)
	defer func() {
		_ = vm.Set("config", goja.Undefined())
		_ = vm.Set("fetch", goja.Undefined())
		_ = vm.Set("nocache", goja.Undefined())
		_ = vm.Set("redirect", goja.Undefined())
		for k := range customContext {
			if !lo.Contains([]string{"cache", "URL", "faker"}, k) {
				_ = vm.Set(k, goja.Undefined())
			}
		}
		gojaVmPool.Put(vm)
	}()
	_ = vm.Set("config", config)
	_ = vm.Set("fetch", helpers.SiteFetch(config))
	if err = vm.Set("nocache", nocache); err != nil {
		log.Println(err)
		return
	}
	_ = vm.Set("redirect", doRedirect)
	for k, v := range customContext {
		_ = vm.Set(k, v)
	}
	var program *goja.Program
	program, err = getJsProgram(name+":cacheKey", string(source)+" cacheKey()")
	if err != nil {
		log.Println(err)
		return
	}
	var v goja.Value
	v, err = vm.RunProgram(program)
	if err != nil {
		log.Println(err)
		return
	}
	hostName := customContext["host"].(string)
	cacheKey := "custom:" + hostName + ":" + name + ":" + v.String()
	program, err = getJsProgram(name+":cacheTtl", string(source)+" cacheTtl()")
	if err != nil {
		log.Println(err)
		return
	}
	v, err = vm.RunProgram(program)
	if err != nil {
		log.Println(err)
		return
	}
	cacheTtl := time.Duration(v.ToInteger()) * time.Second
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
				Path:    "/",
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
		ip := r.Context().Value(types.ContextKeyIp).(string)
		ctx["ip"] = ip

		ctx["country"] = func() string {
			country, _ := geoip.Country(net.ParseIP(ip))
			return country
		}
		ctx["country_group"] = func() types.CountryGroup {
			return internal.DetectCountryGroup(net.ParseIP(ip))
		}
		ctx["redirect_to"] = func(params ...any) {
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
				log.Println("Redirected ", code, url)
			}
		}
		ctx["custom_send"] = func(params ...any) customSendRet {
			if len(params) == 0 {
				return customSendRet{code: http.StatusOK, data: []byte("")}
			}
			var data []byte
			var code = http.StatusOK
			var headers = make(map[string]string)
			if len(params) >= 1 {
				if data1, ok := params[0].([]byte); ok {
					data = data1
				} else {
					data = fmt.Appendf(nil, "%v", params[0])
				}
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
			// Вместо прямой записи, создаем ErrSendResponse и выбрасываем его как исключение Goja
			var h http.Header
			if len(headers) > 0 {
				h = make(http.Header)
				for k, v := range headers {
					h.Set(k, v)
				}
			}
			return customSendRet{data: data, headers: h, code: code}
		}
	}
	addDynamicFunctions(customContext)
	// Adding custom functions to context
	var addCustomFunctions = func(c pongo2.Context) {
		c["add_header"] = func(name, value string) {
			if h, ok := c["_headers"].(http.Header); ok {
				h.Add(name, value)
				c["_headers"] = h
			} else {
				h := http.Header{}
				h.Add(name, value)
				c["_headers"] = h
			}
		}
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
				//GojaVMMutex.Lock(baseName)
				//defer GojaVMMutex.Unlock(baseName)
				vm := gojaVmPool.Get().(*goja.Runtime)
				defer func() {
					_ = vm.Set("config", goja.Undefined())
					_ = vm.Set("fetch", goja.Undefined())
					_ = vm.Set("nocache", goja.Undefined())
					_ = vm.Set("redirect", goja.Undefined())
					for k := range c {
						if !lo.Contains([]string{"cache", "URL", "faker"}, k) {
							_ = vm.Set(k, goja.Undefined())
						}
					}
					for argIndex := range args {
						var argName = fmt.Sprintf("arg%d", argIndex)
						_ = vm.Set(argName, goja.Undefined())
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
				program, err := getJsProgram("function:"+funcName, string(source)+" "+funcName+"("+argsString+")")
				if err != nil {
					log.Println(err)
					return nil
				}
				v, err := vm.RunProgram(program)
				if err != nil {
					log.Println(err)
					return nil
				}
				return v.Export()
			}
		}
	}

	var ctx pongo2.Context
	recreate := func() (parsed []byte, err error) {
		var prepareCtx map[string]interface{}
		prepareCtx, err = func() (prepareCtx map[string]interface{}, err error) {
			// first - run prepare() function
			vm := gojaVmPool.Get().(*goja.Runtime)
			defer func() {
				_ = vm.Set("config", goja.Undefined())
				_ = vm.Set("fetch", goja.Undefined())
				_ = vm.Set("nocache", goja.Undefined())
				_ = vm.Set("redirect", goja.Undefined())
				for k := range customContext {
					if !lo.Contains([]string{"cache", "URL", "faker"}, k) {
						_ = vm.Set(k, goja.Undefined())
					}
				}
				gojaVmPool.Put(vm)
			}()
			_ = vm.Set("config", config)
			_ = vm.Set("fetch", helpers.SiteFetch(config))
			_ = vm.Set("nocache", nocache)
			_ = vm.Set("redirect", doRedirect)
			for k, v := range customContext {
				_ = vm.Set(k, v)
			}
			program, err = getJsProgram(name+":prepare", string(source)+" prepare()")
			if err != nil {
				log.Println(err)
				return
			}
			v, err = vm.RunProgram(program)
			if err != nil {
				log.Println(err)
				return
			}
			if ctx, ok := v.Export().(map[string]any); ok {
				prepareCtx = ctx
			}
			return
		}()
		if err != nil {
			return
		}
		if prepareCtx != nil {
			customContext.Update(prepareCtx)
		}
		ctx = generateContext(name, path, customContext)
		addCustomFunctions(ctx)
		//addDynamicFunctions(ctx)
		vm := gojaVmPool.Get().(*goja.Runtime)
		defer func() {
			_ = vm.Set("config", goja.Undefined())
			_ = vm.Set("fetch", goja.Undefined())
			_ = vm.Set("nocache", goja.Undefined())
			_ = vm.Set("redirect", goja.Undefined())
			for k := range ctx {
				if !lo.Contains([]string{"cache", "URL", "faker"}, k) {
					_ = vm.Set(k, goja.Undefined())
				}
			}
			gojaVmPool.Put(vm)
		}()
		_ = vm.Set("config", config)
		_ = vm.Set("fetch", helpers.SiteFetch(config))
		_ = vm.Set("nocache", nocache)
		_ = vm.Set("redirect", doRedirect)
		for k, val := range ctx {
			_ = vm.Set(k, val)
		}
		program, err = getJsProgram(name+":render", string(source)+" render()")
		if err != nil {
			log.Println(err)
			return
		}
		v, err = vm.RunProgram(program)
		if err != nil {
			// Перехватываем Goja-исключение, если оно является ErrSendResponse
			if gojaErr, ok := err.(*goja.Exception); ok {
				if sendErr, isSendErr := gojaErr.Value().Export().(ErrSendResponse); isSendErr {
					err = sendErr // Пробрасываем нашу кастомную ошибку дальше
					return
				}
			}
			log.Println(err)
			return
		}
		if ret, ok := v.Export().(map[string]any); ok {
			// if render() returns object - output it as json
			e := ErrSendResponse{JSON: ret}
			if h, ok := ctx["_headers"].(http.Header); ok {
				e.Headers = h
			}
			err = e
			return
		}
		if ret, ok := v.Export().(string); ok {
			// if render() returns string - output it as is
			parsed = []byte(ret)
			e := ErrSendResponse{Text: ret}
			if h, ok := ctx["_headers"].(http.Header); ok {
				e.Headers = h
			}
			err = e
			return
		}
		if ret, ok := v.Export().([]byte); ok {
			parsed = ret
			e := ErrSendResponse{Data: ret}
			if h, ok := ctx["_headers"].(http.Header); ok {
				e.Headers = h
			}
			err = e
			return
		}
		if ret, ok := v.Export().(customSendRet); ok {
			e := ErrSendResponse{Data: ret.data, Headers: ret.headers, Code: ret.code}
			err = e
			return
		}
		if ret, ok := v.Export().([]any); ok {
			e := ErrSendResponse{JSON: ret}
			if h, ok := ctx["_headers"].(http.Header); ok {
				e.Headers = h
			}
			err = e
			return
		}
		if ret, ok := v.Export().(redirectRet); ok {
			e := ErrSendResponse{Redirect: ret.url, Code: ret.code}
			if h, ok := ctx["_headers"].(http.Header); ok {
				e.Headers = h
			}
			err = e
			return
		}
		var template *pongo2.Template
		template, err = GetTemplate("custom-"+name, path)
		if err != nil {
			return
		}
		parsed, err = template.ExecuteBytes(ctx)
		if err != nil {
			return
		}
		if config.General.MinifyHtml {
			parsed, err = helpers.MinifyBytes(parsed)
			if err != nil {
				log.Println("can't minify html:", err, name, path, config.Hostname)
			}
		}
		return
	}

	if cacheTtl > 0 {
		if parsed, err = db.GetCachedTimeout(cacheKey, cacheTtl, time.Duration(math.Max(float64(time.Second*5), float64(cacheTtl/2))), recreate, nocache); err != nil {
			return
		}
		c := generateContext(name, path, customContext)
		addCustomFunctions(c)
		//addDynamicFunctions(c)
		parsed = InsertDynamic(parsed, path, c)
		parsed = postHook(parsed, name, path, config, c, nocache)
		return
	}
	if parsed, err = recreate(); err != nil {
		return
	}
	addDynamicFunctions(ctx)
	parsed = InsertDynamic(parsed, path, ctx)
	parsed = postHook(parsed, name, path, config, ctx, nocache)
	return
}
