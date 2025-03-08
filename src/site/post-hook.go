package site

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/flosch/pongo2/v6"
	"github.com/samber/lo"
	"sersh.com/totaltube/frontend/helpers"
	"sersh.com/totaltube/frontend/types"
)

// Function postHook is for changing the data after it has been processed by the template.
func postHook(parsed []byte, name, path string, config *types.Config, c pongo2.Context, nocache bool) []byte {
	matches, _ := filepath.Glob(filepath.Join(path, "extensions/posthook-*.js"))
	for _, m := range matches {
		baseName := filepath.Base(m)
		funcName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(baseName, "posthook-"), ".js"))
		if funcName == "" {
			continue
		}
		func() {
			var source = getJsSource(m)
			if len(source) == 0 {
				return
			}
			vm := gojaVmPool.Get().(*goja.Runtime)
			defer func() {
				_ = vm.Set("config", goja.Undefined())
				_ = vm.Set("fetch", goja.Undefined())
				_ = vm.Set("nocache", goja.Undefined())
				_ = vm.Set("parsed_html", goja.Undefined())
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
			_ = vm.Set("parsed_html", string(parsed))
			for k, v := range c {
				_ = vm.Set(k, v)
			}
			var program *goja.Program
			var err error
			program, err = getJsProgram("posthook:"+funcName, string(source)+" "+funcName+"(parsed_html)")
			if err != nil {
				log.Println(err)
				return
			}
			var v goja.Value
			v, err = vm.RunProgram(program)
			if err != nil {
				log.Println(err, path, name, config.Hostname)
				return
			}
			res := v.String()
			if res != "" {
				parsed = []byte(res)
			}
		}()
	}
	return parsed
}
