package gcache

import (
	"container/list"
	"time"
)

// Discards the least recently used items first.
type LRUCache struct {
	baseCache
	items     map[interface{}]*list.Element
	evictList *list.List
}

func newLRUCache(cb *CacheBuilder) *LRUCache {
	c := &LRUCache{}
	buildCache(&c.baseCache, cb)

	c.evictList = list.New()
	c.items = make(map[interface{}]*list.Element, c.size+1)
	return c
}

func (c *LRUCache) set(key, value interface{}) (interface{}, error) {
	// Check for existing item
	var item *lruItem
	if it, ok := c.items[key]; ok {
		c.evictList.MoveToFront(it)
		item = it.Value.(*lruItem)
		item.value = value
	} else {
		// Verify size not exceeded
		if c.evictList.Len() >= c.size {
			c.evict(1)
		}
		item = &lruItem{
			key:   key,
			value: value,
		}
		c.items[key] = c.evictList.PushFront(item)
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

// set a new key-value pair
func (c *LRUCache) Set(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value)
}

// Get a value from cache pool using key if it exists.
// If it dose not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *LRUCache) Get(key interface{}) (interface{}, error) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		it := item.Value.(*lruItem)
		if !it.IsExpired(nil) {
			c.mu.Lock()
			defer c.mu.Unlock()
			return it.value, nil
		}
		c.mu.Lock()
		c.removeElement(item)
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
	return it.(*lruItem).value, nil
}

// evict removes the oldest item from the cache.
func (c *LRUCache) evict(count int) {
	for i := 0; i < count; i++ {
		ent := c.evictList.Back()
		if ent == nil {
			return
		} else {
			c.removeElement(ent)
		}
	}
}

// Removes the provided key from the cache.
func (c *LRUCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LRUCache) remove(key interface{}) bool {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	entry := e.Value.(*lruItem)
	delete(c.items, entry.key)
	if c.evictedFunc != nil {
		entry := e.Value.(*lruItem)
		go (*c.evictedFunc)(entry.key, entry.value)
	}
}

// Returns a slice of the keys in the cache.
func (c *LRUCache) Keys() []interface{} {
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
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Completely clear the cache
func (c *LRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictList = list.New()
	c.items = make(map[interface{}]*list.Element, c.size)
}

// evict all expired entry
func (c *LRUCache) gc() {
	now := time.Now()
	keys := []interface{}{}
	c.mu.RLock()
	for k, item := range c.items {
		if item.Value.(*lruItem).IsExpired(&now) {
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

type lruItem struct {
	key        interface{}
	value      interface{}
	expiration *time.Time
}

// returns boolean value whether this item is expired or not.
func (it *lruItem) IsExpired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		t := time.Now()
		now = &t
	}
	return it.expiration.Before(*now)
}
