package site

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"

	"sersh.com/totaltube/frontend/db"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

var jsVMs = make(map[string]*goja.Runtime)
var jsMutex sync.Mutex

var jsPrograms = make(map[string]*goja.Program)

var jsSources = make(map[string]struct {
	Source       []byte
	LastModified time.Time
})

func getJsSource(path string) []byte {
	jsMutex.Lock()
	defer jsMutex.Unlock()
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
	n.Source, _ = ioutil.ReadFile(path)
	jsSources[path] = n
	return n.Source
}

var faker = gofakeit.NewCrypto()

func getJsVM(name string) *goja.Runtime {
	jsMutex.Lock()
	defer jsMutex.Unlock()
	if VM, ok := jsVMs[name]; ok {
		return VM
	}
	VM := goja.New()
	var gojaRegistry = new(require.Registry)
	gojaRegistry.Enable(VM)
	console.Enable(VM)
	err := VM.Set("fetch", helpers.Fetch)
	if err != nil {
		panic(err)
	}
	err = VM.Set("cache", func(cacheKey string, timeout string, recreate func() string) string {
		timeoutDuration := types.ParseHumanDuration(timeout)
		extendedTimeout := timeoutDuration / 2
		res, _ := db.GetCachedTimeout(cacheKey, timeoutDuration, extendedTimeout, func() ([]byte, error) {
			recreateRes := recreate()
			return []byte(recreateRes), nil
		}, false)
		return string(res)
	})
	if err != nil {
		panic(err)
	}
	err = VM.Set("faker", faker)
	if err != nil {
		panic(err)
	}
	err = VM.Set("URL", func(call goja.ConstructorCall) *goja.Object {
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
	jsVMs[name] = VM
	return VM
}

func getJsProgram(name string, code string) (program *goja.Program, err error) {
	var ok bool
	hash := helpers.Md5Hash(code)
	if program, ok = jsPrograms[name+":"+hash]; ok {
		return
	}
	program, err = goja.Compile(name, code, true)
	if err != nil {
		return
	}
	jsPrograms[name+":"+hash] = program
	return
}
