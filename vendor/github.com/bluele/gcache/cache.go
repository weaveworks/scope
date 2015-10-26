package gcache

import (
	"errors"
	"github.com/bluele/gcache/singleflight"
	"sync"
	"time"
)

const (
	TYPE_SIMPLE = "simple"
	TYPE_LRU    = "lru"
	TYPE_LFU    = "lfu"
	TYPE_ARC    = "arc"
)

var NotFoundKeyError = errors.New("Not found error.")

type Cache interface {
	Set(interface{}, interface{})
	Get(interface{}) (interface{}, error)
	Remove(interface{}) bool
	Purge()
	Keys() []interface{}
	Len() int
	gc()
}

type baseCache struct {
	size        int
	loaderFunc  *LoaderFunc
	evictedFunc *EvictedFunc
	addedFunc   *AddedFunc
	expiration  *time.Duration
	mu          sync.RWMutex
	loadGroup   singleflight.Group
}

type LoaderFunc func(interface{}) (interface{}, error)

type EvictedFunc func(interface{}, interface{})

type AddedFunc func(interface{}, interface{})

type CacheBuilder struct {
	tp          string
	size        int
	loaderFunc  *LoaderFunc
	evictedFunc *EvictedFunc
	addedFunc   *AddedFunc
	expiration  *time.Duration
	gcInterval  *time.Duration
}

func New(size int) *CacheBuilder {
	if size <= 0 {
		panic("gcache: size <= 0")
	}
	return &CacheBuilder{
		tp:   TYPE_SIMPLE,
		size: size,
	}
}

func (cb *CacheBuilder) LoaderFunc(loaderFunc LoaderFunc) *CacheBuilder {
	cb.loaderFunc = &loaderFunc
	return cb
}

func (cb *CacheBuilder) EnableGC(interval time.Duration) *CacheBuilder {
	cb.gcInterval = &interval
	return cb
}

func (cb *CacheBuilder) EvictType(tp string) *CacheBuilder {
	cb.tp = tp
	return cb
}

func (cb *CacheBuilder) Simple() *CacheBuilder {
	return cb.EvictType(TYPE_SIMPLE)
}

func (cb *CacheBuilder) LRU() *CacheBuilder {
	return cb.EvictType(TYPE_LRU)
}

func (cb *CacheBuilder) LFU() *CacheBuilder {
	return cb.EvictType(TYPE_LFU)
}

func (cb *CacheBuilder) ARC() *CacheBuilder {
	return cb.EvictType(TYPE_ARC)
}

func (cb *CacheBuilder) EvictedFunc(evictedFunc EvictedFunc) *CacheBuilder {
	cb.evictedFunc = &evictedFunc
	return cb
}

func (cb *CacheBuilder) AddedFunc(addedFunc AddedFunc) *CacheBuilder {
	cb.addedFunc = &addedFunc
	return cb
}

func (cb *CacheBuilder) Expiration(expiration time.Duration) *CacheBuilder {
	cb.expiration = &expiration
	return cb
}

func (cb *CacheBuilder) Build() Cache {
	cache := cb.build()
	if cb.gcInterval != nil {
		go func() {
			t := time.NewTicker(*cb.gcInterval)
			for {
				select {
				case <-t.C:
					go cache.gc()
				}
			}
			t.Stop()
		}()
	}
	return cache
}

func (cb *CacheBuilder) build() Cache {
	switch cb.tp {
	case TYPE_SIMPLE:
		return newSimpleCache(cb)
	case TYPE_LRU:
		return newLRUCache(cb)
	case TYPE_LFU:
		return newLFUCache(cb)
	case TYPE_ARC:
		return newARC(cb)
	default:
		panic("gcache: Unknown type " + cb.tp)
	}
}

func buildCache(c *baseCache, cb *CacheBuilder) {
	c.size = cb.size
	c.loaderFunc = cb.loaderFunc
	c.expiration = cb.expiration
	c.addedFunc = cb.addedFunc
	c.evictedFunc = cb.evictedFunc
}

// load a new value using by specified key.
func (c *baseCache) load(key interface{}, cb func(interface{}, error) (interface{}, error)) (interface{}, error) {
	v, err := c.loadGroup.Do(key, func() (interface{}, error) {
		return cb((*c.loaderFunc)(key))
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}
