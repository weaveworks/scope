package endpoint

import (
	"net"
	"strings"
	"time"

	"github.com/bluele/gcache"
)

const (
	rAddrCacheLen        = 500 // Default cache length
	rAddrBacklog         = 1000
	rAddrCacheExpiration = 30 * time.Minute
)

type revResFunc func(addr string) (names []string, err error)

// ReverseResolver is a caching, reverse resolver
type ReverseResolver struct {
	addresses chan string
	cache     gcache.Cache
	Throttle  <-chan time.Time // Made public for mocking
	Resolver  revResFunc
}

// NewReverseResolver starts a new reverse resolver that
// performs reverse resolutions and caches the result.
func NewReverseResolver() *ReverseResolver {
	r := ReverseResolver{
		addresses: make(chan string, rAddrBacklog),
		cache:     gcache.New(rAddrCacheLen).LRU().Expiration(rAddrCacheExpiration).Build(),
		Throttle:  time.Tick(time.Second / 10),
		Resolver:  net.LookupAddr,
	}
	go r.loop()
	return &r
}

// Get the reverse resolution for an IP address if already in the cache,
// a gcache.NotFoundKeyError error otherwise.
// Note: it returns one of the possible names that can be obtained for that IP.
func (r *ReverseResolver) Get(address string) (string, error) {
	val, err := r.cache.Get(address)
	if err == nil {
		return val.(string), nil
	}
	if err == gcache.NotFoundKeyError {
		// we trigger a asynchronous reverse resolution when not cached
		select {
		case r.addresses <- address:
		default:
		}
	}
	return "", err
}

func (r *ReverseResolver) loop() {
	for request := range r.addresses {
		<-r.Throttle // rate limit our DNS resolutions
		// and check if the answer is already in the cache
		if _, err := r.cache.Get(request); err == nil {
			continue
		}
		names, err := r.Resolver(request)
		if err == nil && len(names) > 0 {
			name := strings.TrimRight(names[0], ".")
			r.cache.Set(request, name)
		}
	}
}

// Stop the async reverse resolver
func (r *ReverseResolver) Stop() {
	close(r.addresses)
}
