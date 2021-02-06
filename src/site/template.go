package site

import (
	"github.com/dop251/goja"
	"github.com/flosch/pongo2/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
	"github.com/segmentio/encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
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

func NewTemplates(path string, config *Config) *templates {
	n := templates{path: path, templates: make(map[string]*pongo2.Template)}
	n.templateSet = pongo2.NewSet(filepath.Base(path), pongo2.DefaultLoader)
	n.templateSet.Options.LStripBlocks = true
	n.templateSet.Options.TrimBlocks = true
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
				// ждем сигнала при изменении файлов шаблонов
				info := <-c
				var changedTemplatePath = info.Path()
				abs, _ := filepath.Abs(filepath.Join(path, "templates"))
				if filepath.Dir(changedTemplatePath) == abs && filepath.Ext(changedTemplatePath) == ".twig" {
					templateName := strings.TrimSuffix(filepath.Base(changedTemplatePath), filepath.Ext(changedTemplatePath))
					switch templateName {
					case "top-categories", "category", "top-content", "404", "500",
						"model", "models", "content-item", "search":
						err := db.ClearCacheByPrefix(templateName + ":")
						if err != nil {
							log.Println(err)
						}
					}
					if strings.HasPrefix(templateName, "custom-") {
						err := db.ClearCacheByPrefix("custom:" + strings.TrimPrefix(templateName, "custom-") + ":")
						if err != nil {
							log.Println(err)
						}
					}
				}
				n.Lock()
				n.lastChange = time.Now()
				n.Unlock()
				// Через 1.5 секунды после последнего изменения инвалидируем весь кэш шаблонов
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

func (st *siteTemplatesT) get(name, path string, config *Config) (*pongo2.Template, error) {
	st.Lock()
	if ts, ok := st.siteTemplates[path]; ok {
		st.Unlock()
		return ts.get(name)
	}
	st.siteTemplates[path] = NewTemplates(path, config)
	st.Unlock()
	return st.siteTemplates[path].get(name)
}

var siteTemplates = siteTemplatesT{siteTemplates: map[string]*templates{}}

func GetTemplate(name, path string, config *Config) (*pongo2.Template, error) {
	return siteTemplates.get(name, path, config)
}

// uncachedPrepare - функция, которая подготавливает контекст для незакэшированного шаблона
func ParseTemplate(name, path string, config *Config, customContext pongo2.Context,
	nocache bool, cacheKey string, cacheTtl time.Duration,
	uncachedPrepare func(ctx pongo2.Context) (pongo2.Context, error)) (parsed []byte, err error) {
	if !nocache {
		cached := db.GetCached(cacheKey)
		if cached != nil {
			c := generateContext(name, path, customContext)
			parsed, err = InsertDynamic(cached, c)
			return
		}
	}
	customContext, err = uncachedPrepare(customContext)
	if err != nil {
		return
	}
	c := generateContext(name, path, customContext)
	var template *pongo2.Template
	template, err = GetTemplate(name, path, config)
	if err != nil {
		return
	}
	parsed, err = template.ExecuteBytes(c)
	if err != nil {
		return
	}
	if config.General.MinifyHtml {
		parsed = helpers.MinifyBytes(parsed)
	}
	err = db.PutCached(cacheKey, parsed, cacheTtl)
	if err != nil {
		log.Println("can't put item in cache: ", err)
	}
	parsed, err = InsertDynamic(parsed, c)
	return
}

type redirectRet struct {
	url string
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
	var source []byte
	extensionFile := filepath.Join(path, "extensions/template-"+name+".js")
	source, err = ioutil.ReadFile(extensionFile)
	if err != nil {
		err = errors.Wrap(err, "can't open extensionFile "+extensionFile)
	}
	VM := getJsVM(name)
	VM.Set("config", config)
	VM.Set("nocache", nocache)
	VM.Set("redirect", doRedirect)
	for k, v := range customContext {
		VM.Set(k, v)
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
	cacheKey := "custom:" + name + ":" + v.String()
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
	if cacheTtl <= 0 {
		nocache = true
	}
	if !nocache {
		cached := db.GetCached(cacheKey)
		if cached != nil {
			c := generateContext(name, path, customContext)
			parsed, err = InsertDynamic(cached, c)
			return
		}
	}
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
	ctx := generateContext(name, path, customContext)
	program, err = getJsProgram(name+":render", string(source) +" render()")
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
	template, err = GetTemplate("custom-"+name, path, config)
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
	err = db.PutCached(cacheKey, parsed, cacheTtl)
	if err != nil {
		log.Println("can't put item in cache: ", err)
	}
	parsed, err = InsertDynamic(parsed, ctx)
	return
}
