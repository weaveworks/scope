package gcache

import (
	"time"
)

// SimpleCache has no clear priority for evict cache. It depends on key-value map order.
type SimpleCache struct {
	baseCache
	items map[interface{}]*simpleItem
}

func newSimpleCache(cb *CacheBuilder) *SimpleCache {
	c := &SimpleCache{}
	buildCache(&c.baseCache, cb)

	c.items = make(map[interface{}]*simpleItem, c.size)
	return c
}

// set a new key-value pair
func (c *SimpleCache) Set(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value)
}

func (c *SimpleCache) set(key, value interface{}) (interface{}, error) {
	// Check for existing item
	item, ok := c.items[key]
	if ok {
		item.value = value
	} else {
		// Verify size not exceeded
		if len(c.items) >= c.size {
			c.evict(1)
		}
		item = &simpleItem{
			value: value,
		}
		c.items[key] = item
	}

	if c.expiration != nil {
		t := time.Now().Add(*c.expiration)
		item.expiration = &t
	}

	if c.addedFunc != nil {
		go (*c.addedFunc)(key, value)
	}

	return item, nil
}

// Get a value from cache pool using key if it exists.
// If it dose not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *SimpleCache) Get(key interface{}) (interface{}, error) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		if !item.IsExpired(nil) {
			return item.value, nil
		}
		c.mu.Lock()
		c.remove(key)
		c.mu.Unlock()
	}

	if c.loaderFunc == nil {
		return nil, NotFoundKeyError
	}

	it, err := c.load(key, func(v interface{}, e error) (interface{}, error) {
		if e == nil {
			c.mu.Lock()
			defer c.mu.Unlock()
			return c.set(key, v)
		}
		return nil, e
	})
	if err != nil {
		return nil, err
	}
	return it.(*simpleItem).value, nil
}

func (c *SimpleCache) evict(count int) {
	now := time.Now()
	current := 0
	for key, item := range c.items {
		if current >= count {
			return
		}
		if item.expiration == nil || now.After(*item.expiration) {
			defer c.remove(key)
			current += 1
		}
	}
}

// Removes the provided key from the cache.
func (c *SimpleCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *SimpleCache) remove(key interface{}) bool {
	item, ok := c.items[key]
	if ok {
		delete(c.items, key)
		if c.evictedFunc != nil {
			go (*c.evictedFunc)(key, item.value)
		}
		return true
	}
	return false
}

// Returns a slice of the keys in the cache.
func (c *SimpleCache) Keys() []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]interface{}, len(c.items))
	i := 0
	for k := range c.items {
		keys[i] = k
		i++
	}

	return keys
}

// Returns the number of items in the cache.
func (c *SimpleCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Completely clear the cache
func (c *SimpleCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[interface{}]*simpleItem, c.size)
}

// evict all expired entry
func (c *SimpleCache) gc() {
	now := time.Now()
	keys := []interface{}{}
	c.mu.RLock()
	for k, item := range c.items {
		if item.IsExpired(&now) {
			keys = append(keys, k)
		}
	}
	c.mu.RUnlock()
	if len(keys) == 0 {
		return
	}
	c.mu.Lock()
	for _, k := range keys {
		c.remove(k)
	}
	c.mu.Unlock()
}

type simpleItem struct {
	value      interface{}
	expiration *time.Time
}

// returns boolean value whether this item is expired or not.
func (si *simpleItem) IsExpired(now *time.Time) bool {
	if si.expiration == nil {
		return false
	}
	if now == nil {
		t := time.Now()
		now = &t
	}
	return si.expiration.Before(*now)
}
