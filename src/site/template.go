package site

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
	"github.com/segmentio/encoding/json"
	"log"
	"math"
	"path/filepath"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/internal"
	"sersh.com/totaltube/frontend/types"
	"strings"
	"sync"
	"time"
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
							"model", "models", "content-item", "search", "fake-player":
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

func ParseTemplate(name, path string, config *Config, customContext pongo2.Context,
	nocache bool, cacheKey string, cacheTtl time.Duration,
	uncachedPrepare func(ctx pongo2.Context) (pongo2.Context, error)) (parsed []byte, err error) {
	var c pongo2.Context
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
	var cached []byte
	cached, err = db.GetCachedTimeout(cacheKey, cacheTtl, time.Duration(math.Max(float64(time.Minute*10), float64(cacheTtl))), func() (result []byte, err error) {
		var c pongo2.Context
		c, err = uncachedPrepare(customContext)
		if err != nil {
			log.Println(err)
			return
		}
		c = generateContext(name, path, c)
		addCustomFunctions(c)
		var template *pongo2.Template
		template, err = GetTemplate(name, path)
		if err != nil {
			return
		}
		result, err = template.ExecuteBytes(c)
		if err != nil {
			return
		}
		if config.General.MinifyHtml {
			result = helpers.MinifyBytes(result)
		}
		return
	}, nocache)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			log.Println(err)
		}
		return
	}
	c = generateContext(name, path, customContext)
	addCustomFunctions(c)
	parsed, err = InsertDynamic(cached, c)
	return
}

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
func ParseCustomTemplate(name, path string, config *Config,
	customContext pongo2.Context, nocache bool, c *fiber.Ctx) (parsed []byte, err error) {
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
			parsed, err = json.Marshal(ret)
			if err != nil {
				log.Println(err)
			}
			c.Set(fiber.HeaderContentType, "application/json")
			err = c.Send(parsed)
			if err != nil {
				return
			}
			err = types.ErrResponseSent
			return
		}
		if ret, ok := v.Export().(string); ok {
			// if render() returns string - output it as is
			parsed = []byte(ret)
			return
		}
		if ret, ok := v.Export().(redirectRet); ok {
			err = c.Redirect(ret.url, ret.code)
			if err != nil {
				log.Println(err)
				return
			}
			err = types.ErrResponseSent
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
		parsed, err = InsertDynamic(parsed, c)
		return
	}
	if parsed, err = recreate(); err != nil {
		return
	}
	parsed, err = InsertDynamic(parsed, ctx)
	addCustomFunctions(ctx)
	return
}
