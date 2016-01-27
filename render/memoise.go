package render

import (
	"fmt"
	"reflect"

	"github.com/bluele/gcache"

	"github.com/weaveworks/scope/report"
)

var renderCache = gcache.New(100).LRU().Build()

type memoise struct {
	Renderer
}

// Memoise wraps the renderer in a loving embrace of caching
func Memoise(r Renderer) Renderer { return &memoise{r} }

// Render produces a set of RenderableNodes given a Report.
// Ideally, it just retrieves it from the cache, otherwise it calls through to
// `r` and stores the result.
func (m *memoise) Render(rpt report.Report) RenderableNodes {
	key := ""
	v := reflect.ValueOf(m.Renderer)
	switch v.Kind() {
	case reflect.Ptr, reflect.Func:
		key = fmt.Sprintf("%s-%x", rpt.ID, v.Pointer())
	default:
		return m.Renderer.Render(rpt)
	}
	if result, err := renderCache.Get(key); err == nil {
		return result.(RenderableNodes)
	}
	output := m.Renderer.Render(rpt)
	renderCache.Set(key, output)
	return output
}

// ResetCache blows away the rendered node cache.
func ResetCache() {
	renderCache.Purge()
}
