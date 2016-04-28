package render

import (
	"fmt"
	"math/rand"

	"github.com/bluele/gcache"

	"github.com/weaveworks/scope/report"
)

var renderCache = gcache.New(100).LRU().Build()

type memoise struct {
	Renderer
	id string
}

// Memoise wraps the renderer in a loving embrace of caching
func Memoise(r Renderer) Renderer {
	return &memoise{
		Renderer: r,
		id:       fmt.Sprintf("%x", rand.Int63()),
	}
}

// Render produces a set of Nodes given a Report.
// Ideally, it just retrieves it from the cache, otherwise it calls through to
// `r` and stores the result.
func (m *memoise) Render(rpt report.Report, dct Decorator) report.Nodes {
	key := fmt.Sprintf("%s-%s", rpt.ID, m.id)
	if dct == nil {
		if result, err := renderCache.Get(key); err == nil {
			return result.(report.Nodes)
		}
	}
	output := m.Renderer.Render(rpt, dct)
	if dct == nil {
		renderCache.Set(key, output)
	}
	return output
}

// ResetCache blows away the rendered node cache.
func ResetCache() {
	renderCache.Purge()
}
