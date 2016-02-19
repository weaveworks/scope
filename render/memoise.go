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

// Render produces a set of RenderableNodes given a Report.
// Ideally, it just retrieves it from the cache, otherwise it calls through to
// `r` and stores the result.
func (m *memoise) Render(rpt report.Report) RenderableNodes {
	key := fmt.Sprintf("%s-%s", rpt.ID, m.id)
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
