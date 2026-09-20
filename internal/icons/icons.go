package icons

import (
	"embed"
	"strings"
	"sync"
)

//go:embed builtin/*.svg
var builtinFS embed.FS

var (
	cache   map[string]string
	loadMu  sync.Once
)

func loadAll() {
	cache = make(map[string]string)
	entries, err := builtinFS.ReadDir("builtin")
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".svg") {
			continue
		}
		data, err := builtinFS.ReadFile("builtin/" + e.Name())
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".svg")
		cache[name] = string(data)
	}
}

func Get(name string) string {
	loadMu.Do(loadAll)
	if svg, ok := cache[name]; ok {
		return svg
	}
	if svg, ok := cache["generic"]; ok {
		return svg
	}
	return ""
}

func Has(name string) bool {
	loadMu.Do(loadAll)
	_, ok := cache[name]
	return ok
}

func List() []string {
	loadMu.Do(loadAll)
	names := make([]string, 0, len(cache))
	for name := range cache {
		names = append(names, name)
	}
	return names
}
