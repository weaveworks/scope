package render

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/bluele/gcache"

	"github.com/weaveworks/scope/report"
)

// renderCache is keyed on the combination of Memoiser and report
// id. It contains promises of report.Nodes, which result from
// rendering the report with the Memoiser's renderer.
//
// The use of promises ensures that in the absence of cache evictions
// a memoiser will only ever render a report once, even when Render()
// is invoked concurrently.
var renderCache = gcache.New(100).LRU().Build()

type memoise struct {
	sync.Mutex
	Renderer
	id string
}

// Memoise wraps the renderer in a loving embrace of caching.
func Memoise(r Renderer) Renderer {
	return &memoise{
		Renderer: r,
		id:       fmt.Sprintf("%x", rand.Int63()),
	}
}

// Render produces a set of Nodes given a Report.  Ideally, it just
// retrieves a promise from the cache and returns its value, otherwise
// it stores a new promise and fulfils it by calling through to
// m.Renderer.
//
// The cache is bypassed when rendering a report with a decorator.
func (m *memoise) Render(rpt report.Report, dct Decorator) Nodes {
	if dct != nil {
		return m.Renderer.Render(rpt, dct)
	}

	key := fmt.Sprintf("%s-%s", rpt.ID, m.id)

	m.Lock()
	v, err := renderCache.Get(key)
	if err == nil {
		m.Unlock()
		return v.(*promise).Get()
	}
	promise := newPromise()
	renderCache.Set(key, promise)
	m.Unlock()

	output := m.Renderer.Render(rpt, dct)

	promise.Set(output)

	return output
}

type promise struct {
	val  Nodes
	done chan struct{}
}

func newPromise() *promise {
	return &promise{done: make(chan struct{})}
}

func (p *promise) Set(val Nodes) {
	p.val = val
	close(p.done)
}

func (p *promise) Get() Nodes {
	<-p.done
	return p.val
}

// ResetCache blows away the rendered node cache.
func ResetCache() {
	renderCache.Purge()
}
