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

type revResRequest struct {
	address string
	done    chan struct{}
}

// ReverseResolver is a caching, reverse resolver
type reverseResolver struct {
	addresses chan revResRequest
	cache     gcache.Cache
	resolver  revResFunc
}

// NewReverseResolver starts a new reverse resolver that
// performs reverse resolutions and caches the result.
func newReverseResolver() *reverseResolver {
	r := reverseResolver{
		addresses: make(chan revResRequest, rAddrBacklog),
		cache:     gcache.New(rAddrCacheLen).LRU().Expiration(rAddrCacheExpiration).Build(),
		resolver:  net.LookupAddr,
	}
	go r.loop()
	return &r
}

// Get the reverse resolution for an IP address if already in the cache,
// a gcache.NotFoundKeyError error otherwise.
// Note: it returns one of the possible names that can be obtained for that IP.
func (r *reverseResolver) Get(address string, wait bool) (string, error) {
	val, err := r.cache.Get(address)
	if err == nil {
		return val.(string), nil
	}
	if err == gcache.NotFoundKeyError {
		request := revResRequest{address: address, done: make(chan struct{})}
		// we trigger a asynchronous reverse resolution when not cached
		select {
		case r.addresses <- request:
			if wait {
				<-request.done
			}
		default:
		}
	}
	return "", err
}

func (r *reverseResolver) loop() {
	throttle := time.Tick(time.Second / 10)
	for request := range r.addresses {
		<-throttle // rate limit our DNS resolutions
		// and check if the answer is already in the cache
		if _, err := r.cache.Get(request.address); err == nil {
			continue
		}
		names, err := r.resolver(request.address)
		if err == nil && len(names) > 0 {
			name := strings.TrimRight(names[0], ".")
			r.cache.Set(request.address, name)
		}
		close(request.done)
	}
}

// Stop the async reverse resolver
func (r *reverseResolver) Stop() {
	close(r.addresses)
}
