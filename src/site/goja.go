package site

import (
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/dop251/goja/parser"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

var gojaVmPool = sync.Pool{
	New: func() interface{} {
		vm := goja.New()
		vm.SetParserOptions(parser.WithDisableSourceMaps)
		var gojaRegistry = new(require.Registry)
		gojaRegistry.Enable(vm)
		console.Enable(vm)
		var err error
		err = vm.Set("cache", func(cacheKey string, timeout string, recreate func() string) string {
			timeoutDuration := types.ParseHumanDuration(timeout)
			var res []byte
			res, err = db.GetCachedTimeout(cacheKey, timeoutDuration, 0, func() (res []byte, err error) {
				defer func() {
					if err1 := recover(); err1 != nil {
						log.Println(err1)
						res = []byte{}
						err = err1.(error)
					}
				}()
				recreateRes := recreate()
				res = []byte(recreateRes)
				return
			}, false)
			if err != nil {
				log.Println(err)
				return ""
			}
			return string(res)
		})
		if err != nil {
			panic(err)
		}
		err = vm.Set("faker", faker)
		if err != nil {
			panic(err)
		}
		err = vm.Set("URL", func(call goja.ConstructorCall) *goja.Object {
			u, _ := url.Parse(call.Argument(0).String())
			if u == nil {
				return nil
			}
			_ = call.This.Set("pathname", u.Path)
			_ = call.This.Set("host", u.Host)
			_ = call.This.Set("hostname", u.Hostname())
			_ = call.This.Set("href", u.String())
			_ = call.This.Set("port", u.Port())
			_ = call.This.Set("protocol", u.Scheme+":")
			_ = call.This.Set("search", u.RawQuery)
			return nil
		})
		return vm
	},
}

var jsSourceMutex sync.Mutex
var jsProgramMutex sync.Mutex

var jsPrograms = make(map[string]*goja.Program)

var jsSources = make(map[string]struct {
	Source       []byte
	LastModified time.Time
})

func getJsSource(path string) []byte {
	jsSourceMutex.Lock()
	defer jsSourceMutex.Unlock()
	if existing, ok := jsSources[path]; ok {
		if info, err := os.Stat(path); err != nil {
			log.Println("can't open", path, ":", err)
			return []byte{}
		} else if !info.ModTime().After(existing.LastModified) {
			return existing.Source
		}
	}
	var n = struct {
		Source       []byte
		LastModified time.Time
	}{}
	if info, err := os.Stat(path); err != nil {
		log.Println("can't open", path, ":", err)
		return []byte{}
	} else {
		n.LastModified = info.ModTime()
	}
	n.Source, _ = os.ReadFile(path)
	jsSources[path] = n
	return n.Source
}

var faker = gofakeit.NewCrypto()

func getJsProgram(name string, code string) (program *goja.Program, err error) {
	jsProgramMutex.Lock()
	defer jsProgramMutex.Unlock()
	var ok bool
	hash := helpers.Md5Hash(code)
	if program, ok = jsPrograms[name+":"+hash]; ok {
		return
	}
	program, err = goja.Compile(name, code, true)
	if err != nil {
		log.Println("Failed to compile JS program:", err)
		return
	}
	jsPrograms[name+":"+hash] = program
	return
}
