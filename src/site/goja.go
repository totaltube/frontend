package site

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"sersh.com/totaltube/frontend/helpers"
	"sync"
)

var jsVMs = make(map[string]*goja.Runtime)
var jsMutex sync.Mutex

var jsPrograms = make(map[string]*goja.Program)

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
	VM.Set("fetch", helpers.Fetch)
	jsVMs[name] = VM
	return VM
}

func getJsProgram(name string, code string) (program *goja.Program, err error) {
	var ok bool
	if program, ok = jsPrograms[name]; ok {
		return
	}
	program, err =  goja.Compile(name, code, true)
	if err != nil {
		return
	}
	jsPrograms[name] = program
	return
}
