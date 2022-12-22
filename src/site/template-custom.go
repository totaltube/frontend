package site

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/flosch/pongo2/v4"
	"github.com/pkg/errors"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
)

type redirectRet struct {
	url  string
	code int
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
	Redirect     string
	RedirectCode int
	JSON         interface{}
	Text         string
	Data         []byte
	Headers      http.Header
}

func (e ErrSendResponse) Error() string {
	return "custom response"
}

func ParseCustomTemplate(name, path string, config *Config,
	customContext pongo2.Context, nocache bool, w http.ResponseWriter, r *http.Request) (parsed []byte, err error) {
	extensionFile := filepath.Join(path, "extensions/route-"+name+".js")
	var source = getJsSource(extensionFile)
	if len(source) == 0 {
		err = errors.New("source template " + extensionFile + " is empty or not exists")
		return
	}
	helpers.KeyMutex.Lock(name)
	defer helpers.KeyMutex.Unlock(name)
	VM := getJsVM(name)
	if err = VM.Set("config", config); err != nil {
		log.Println(err)
		return
	}
	if err = VM.Set("nocache", nocache); err != nil {
		log.Println(err)
		return
	}
	if err = VM.Set("redirect", doRedirect); err != nil {
		log.Println(err)
		return
	}
	for k, v := range customContext {
		if err = VM.Set(k, v); err != nil {
			log.Println(err)
			return
		}
	}
	var program *goja.Program
	program, err = getJsProgram(name+":cacheKey", string(source)+" cacheKey()")
	if err != nil {
		log.Println(err)
		return
	}
	var v goja.Value
	v, err = VM.RunProgram(program)
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
	v, err = VM.RunProgram(program)
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
	}
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
				helpers.KeyMutex.Lock(baseName)
				defer helpers.KeyMutex.Unlock(baseName)
				//GojaVMMutex.Lock(baseName)
				//defer GojaVMMutex.Unlock(baseName)
				VM := getJsVM(baseName)
				_ = VM.Set("config", config)
				_ = VM.Set("nocache", nocache)
				for k, v := range c {
					_ = VM.Set(k, v)
				}
				var argsString string
				var argsNameArray = make([]string, 0, len(args))
				for argIndex, arg := range args {
					var argName = fmt.Sprintf("arg%d", argIndex)
					_ = VM.Set(argName, arg)
					argsNameArray = append(argsNameArray, argName)
				}
				argsString = strings.Join(argsNameArray, ",")
				program, err := getJsProgram("function:"+funcName, string(source)+" "+funcName+"("+argsString+")")
				if err != nil {
					log.Println(err)
					return nil
				}
				v, err := VM.RunProgram(program)
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
		program, err = getJsProgram(name+":prepare", string(source)+" prepare()")
		if err != nil {
			log.Println(err)
			return
		}
		v, err = VM.RunProgram(program)
		if err != nil {
			log.Println(err)
			return
		}
		if ctx, ok := v.Export().(map[string]interface{}); ok {
			customContext.Update(ctx)
		}
		ctx = generateContext(name, path, customContext)
		addCustomFunctions(ctx)
		for k, val := range ctx {
			_ = VM.Set(k, val)
		}
		program, err = getJsProgram(name+":render", string(source)+" render()")
		if err != nil {
			log.Println(err)
			return
		}
		v, err = VM.RunProgram(program)
		if err != nil {
			log.Println(err)
			return
		}
		if ret, ok := v.Export().(map[string]interface{}); ok {
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
		if ret, ok := v.Export().(redirectRet); ok {
			e := ErrSendResponse{Redirect: ret.url, RedirectCode: ret.code}
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
			parsed = helpers.MinifyBytes(parsed)
		}
		return
	}

	if cacheTtl > 0 {
		if parsed, err = db.GetCachedTimeout(cacheKey, cacheTtl, time.Duration(math.Max(float64(time.Second*5), float64(cacheTtl/2))), recreate, nocache); err != nil {
			return
		}
		c := generateContext(name, path, customContext)
		addCustomFunctions(c)
		addDynamicFunctions(c)
		parsed, err = InsertDynamic(parsed, path, c)
		return
	}
	if parsed, err = recreate(); err != nil {
		return
	}
	addDynamicFunctions(ctx)
	parsed, err = InsertDynamic(parsed, path, ctx)
	return
}
