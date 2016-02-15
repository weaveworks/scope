package render

import (
	"github.com/bluele/gcache"

	"github.com/weaveworks/scope/report"
)

type memoise struct {
	Renderer
	cache gcache.Cache
}

// Memoise wraps the renderer in a loving embrace of caching
func Memoise(r Renderer) Renderer { return &memoise{r, gcache.New(10).LRU().Build()} }

// Render produces a set of RenderableNodes given a Report.
// Ideally, it just retrieves it from the cache, otherwise it calls through to
// `r` and stores the result.
func (m *memoise) Render(rpt report.Report) RenderableNodes {
	if result, err := m.cache.Get(rpt.ID); err == nil {
		return result.(RenderableNodes)
	}
	output := m.Renderer.Render(rpt)
	m.cache.Set(rpt.ID, output)
	return output
}

// ResetCache blows away the rendered node cache.
func (m *memoise) ResetCache() {
	m.cache.Purge()
	m.Renderer.ResetCache()
}
