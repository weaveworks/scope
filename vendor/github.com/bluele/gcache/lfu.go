package gcache

import (
	"container/list"
	"time"
)

// Discards the least frequently used items first.
type LFUCache struct {
	baseCache
	items    map[interface{}]*lfuItem
	freqList *list.List // list for freqEntry
}

func newLFUCache(cb *CacheBuilder) *LFUCache {
	c := &LFUCache{}
	buildCache(&c.baseCache, cb)

	c.freqList = list.New()
	c.items = make(map[interface{}]*lfuItem, c.size+1)
	c.freqList.PushFront(&freqEntry{
		freq:  0,
		items: make(map[*lfuItem]byte),
	})
	return c
}

// set a new key-value pair
func (c *LFUCache) Set(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value)
}

func (c *LFUCache) set(key, value interface{}) (*lfuItem, error) {
	// Check for existing item
	item, ok := c.items[key]
	if ok {
		item.value = value
	} else {
		// Verify size not exceeded
		if len(c.items) >= c.size {
			c.evict(1)
		}
		item = &lfuItem{
			key:         key,
			value:       value,
			freqElement: nil,
		}
		el := c.freqList.Front()
		fe := el.Value.(*freqEntry)
		fe.items[item] = 1

		item.freqElement = el
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
func (c *LFUCache) Get(key interface{}) (interface{}, error) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		if !item.IsExpired(nil) {
			c.mu.Lock()
			c.increment(item)
			c.mu.Unlock()
			return item.value, nil
		}
		c.mu.Lock()
		c.removeItem(item)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	li := it.(*lfuItem)
	c.increment(li)
	return li.value, nil
}

func (c *LFUCache) increment(item *lfuItem) {
	currentFreqElement := item.freqElement
	currentFreqEntry := currentFreqElement.Value.(*freqEntry)
	nextFreq := currentFreqEntry.freq + 1
	delete(currentFreqEntry.items, item)

	nextFreqElement := currentFreqElement.Next()
	if nextFreqElement == nil {
		nextFreqElement = c.freqList.InsertAfter(&freqEntry{
			freq:  nextFreq,
			items: make(map[*lfuItem]byte),
		}, currentFreqElement)
	}
	nextFreqElement.Value.(*freqEntry).items[item] = 1
	item.freqElement = nextFreqElement
}

// evict removes the least frequence item from the cache.
func (c *LFUCache) evict(count int) {
	entry := c.freqList.Front()
	for i := 0; i < count; {
		if entry == nil {
			return
		} else {
			for item, _ := range entry.Value.(*freqEntry).items {
				if i >= count {
					return
				}
				c.removeItem(item)
				i++
			}
			entry = entry.Next()
		}
	}
}

// Removes the provided key from the cache.
func (c *LFUCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LFUCache) remove(key interface{}) bool {
	if item, ok := c.items[key]; ok {
		c.removeItem(item)
		return true
	}
	return false
}

// removeElement is used to remove a given list element from the cache
func (c *LFUCache) removeItem(item *lfuItem) {
	delete(c.items, item.key)
	delete(item.freqElement.Value.(*freqEntry).items, item)
	if c.evictedFunc != nil {
		go (*c.evictedFunc)(item.key, item.value)
	}
}

// Returns a slice of the keys in the cache.
func (c *LFUCache) Keys() []interface{} {
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
func (c *LFUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Completely clear the cache
func (c *LFUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.freqList = list.New()
	c.items = make(map[interface{}]*lfuItem, c.size)
}

// evict all expired entry
func (c *LFUCache) gc() {
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

type freqEntry struct {
	freq  uint
	items map[*lfuItem]byte
}

type lfuItem struct {
	key         interface{}
	value       interface{}
	freqElement *list.Element
	expiration  *time.Time
}

// returns boolean value whether this item is expired or not.
func (it *lfuItem) IsExpired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		t := time.Now()
		now = &t
	}
	return it.expiration.Before(*now)
}
