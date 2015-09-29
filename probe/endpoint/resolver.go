package endpoint

import (
	"fmt"
	"net"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	rAddrCacheLen        = 500              // default cache length
	rAddrCacheExpiration = 30 * time.Minute // hard, upper TTL limit
	rAddrBacklog         = 1000
)

var errNotFound = fmt.Errorf("Not found")

type resolverCacheEntry struct {
	name     string
	creation time.Time
}

func newResolverCacheEntry(name string) *resolverCacheEntry {
	return &resolverCacheEntry{
		name:     name,
		creation: time.Now(),
	}
}

type revResFunc func(addr string) (names []string, err error)

// ReverseResolver is a caching, reverse resolver.
type ReverseResolver struct {
	addresses chan string
	cache     *lru.Cache
	Throttle  <-chan time.Time // Made public for mocking
	Resolver  revResFunc
	quit      chan struct{}
}

// NewReverseResolver starts a new reverse resolver that performs reverse
// resolutions and caches the result.
func NewReverseResolver() *ReverseResolver {
	cache, err := lru.New(rAddrCacheLen)
	if err != nil {
		panic("could not create cache")
	}

	r := ReverseResolver{
		addresses: make(chan string, rAddrBacklog),
		cache:     cache,
		Throttle:  time.Tick(time.Second / 10),
		Resolver:  net.LookupAddr,
		quit:      make(chan struct{}),
	}
	go r.loop()
	return &r
}

// Get the reverse resolution for an IP address if already in the cache, a
// gcache.NotFoundKeyError error otherwise. Note: it returns one of the
// possible names that can be obtained for that IP.
func (r *ReverseResolver) Get(address string) (string, error) {
	val, found := r.cache.Get(address)
	if !found {
		// We trigger a asynchronous reverse resolution when not cached
		select {
		case r.addresses <- address:
		default:
		}
		return "", errNotFound
	}
	if _, ok := val.(struct{}); ok {
		return "", errNotFound
	}
	entry := val.(*resolverCacheEntry)
	return entry.name, nil
}

func (r *ReverseResolver) loop() {
	purge := time.Tick(rAddrCacheExpiration)

	for {
		select {
		case request := <-r.addresses:
			// check if the answer is already in the cache
			if _, ok := r.cache.Get(request); ok {
				continue
			}
			<-r.Throttle // rate limit our DNS resolutions
			names, err := r.Resolver(request)
			if err == nil && len(names) > 0 {
				name := strings.TrimRight(names[0], ".")
				r.cache.Add(request, newResolverCacheEntry(name))
			} else {
				r.cache.Add(request, struct{}{})
			}

		case <-purge:
			timeLimit := time.Now().Add(-rAddrCacheExpiration)
			for _, key := range r.cache.Keys() {
				e, _ := r.cache.Peek(key)
				entry, ok := e.(*resolverCacheEntry)
				if !ok {
					panic(fmt.Sprintf("unknown entry type in cache: %+v", e))
				}
				if entry.creation.Before(timeLimit) {
					r.cache.Remove(key)
				}
			}

		case <-r.quit:
			return
		}
	}
}

// Stop the async reverse resolver.
func (r *ReverseResolver) Stop() {
	close(r.quit)
}
