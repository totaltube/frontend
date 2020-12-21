package site

import (
	"github.com/flosch/pongo2/v4"
	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
	"log"
	"path/filepath"
	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
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
			// Нашли наш шаблон
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
				<-c
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
			c := generateContext(name, path, config, customContext)
			parsed, err = InsertDynamic(cached, c)
			return
		}
	}
	customContext, err = uncachedPrepare(customContext)
	if err != nil {
		return
	}
	c := generateContext(name, path, config, customContext)
	var template *pongo2.Template
	template, err = GetTemplate(name, path, config)
	if err != nil {
		return
	}
	parsed, err = template.ExecuteBytes(c)
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
