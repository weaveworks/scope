package gcache

import (
	"container/list"
	"time"
)

// Constantly balances between LRU and LFU, to improve the combined result.
type ARC struct {
	baseCache
	items map[interface{}]*arcItem

	part int
	t1   *arcList
	t2   *arcList
	b1   *arcList
	b2   *arcList
}

func newARC(cb *CacheBuilder) *ARC {
	c := &ARC{
		items: make(map[interface{}]*arcItem),
		t1:    newARCList(),
		t2:    newARCList(),
		b1:    newARCList(),
		b2:    newARCList(),
	}
	buildCache(&c.baseCache, cb)
	return c
}

func (c *ARC) replace(key interface{}) {
	var old interface{}
	if (c.t1.Len() > 0 && c.b2.Has(key) && c.t1.Len() == c.part) || (c.t1.Len() > c.part) {
		old = c.t1.RemoveTail()
		c.b1.PushFront(old)
	} else {
		old = c.t2.RemoveTail()
		c.b2.PushFront(old)
	}
	item, ok := c.items[old]
	if ok {
		delete(c.items, old)
		if c.evictedFunc != nil {
			go (*c.evictedFunc)(item.key, item.value)
		}
	}
}

func (c *ARC) Set(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value)
}

func (c *ARC) set(key, value interface{}) (interface{}, error) {
	item, ok := c.items[key]
	if ok {
		item.value = value
	} else {
		item = &arcItem{
			key:   key,
			value: value,
		}
		c.items[key] = item
	}

	if c.expiration != nil {
		t := time.Now().Add(*c.expiration)
		item.expiration = &t
	}

	if elt := c.b1.Lookup(key); elt != nil {
		c.part = minInt(c.size, c.part+maxInt(c.b2.Len()/c.b1.Len(), 1))
		c.replace(key)
		c.b1.Remove(key, elt)
		c.t2.PushFront(key)
		return item, nil
	}

	if elt := c.b2.Lookup(key); elt != nil {
		c.part = maxInt(0, c.part-maxInt(c.b1.Len()/c.b2.Len(), 1))
		c.replace(key)
		c.b2.Remove(key, elt)
		c.t2.PushFront(key)
		return item, nil
	}

	if c.t1.Len()+c.b1.Len() == c.size {
		if c.t1.Len() < c.size {
			c.b1.RemoveTail()
			c.replace(key)
		} else {
			pop := c.t1.RemoveTail()
			item, ok := c.items[pop]
			if ok {
				delete(c.items, pop)
				if c.evictedFunc != nil {
					go (*c.evictedFunc)(item.key, item.value)
				}
			}
		}
	} else {
		total := c.t1.Len() + c.b1.Len() + c.t2.Len() + c.b2.Len()
		if total >= c.size {
			if total == (2 * c.size) {
				c.b2.RemoveTail()
			}
			c.replace(key)
		}
	}

	c.t1.PushFront(key)

	if c.addedFunc != nil {
		go (*c.addedFunc)(key, value)
	}

	return item, nil
}

// Get a value from cache pool using key if it exists. If not exists and it has LoaderFunc, it will generate the value using you have specified LoaderFunc method returns value.
func (c *ARC) Get(key interface{}) (interface{}, error) {
	rl := false
	c.mu.RLock()
	if elt := c.t1.Lookup(key); elt != nil {
		c.mu.RUnlock()
		rl = true
		c.mu.Lock()
		c.t1.Remove(key, elt)
		item := c.items[key]
		if !item.IsExpired(nil) {
			c.t2.PushFront(key)
			c.mu.Unlock()
			return item.value, nil
		}
		c.b2.PushFront(key)
		if c.evictedFunc != nil {
			go (*c.evictedFunc)(key, elt.Value)
		}
		c.mu.Unlock()
	}
	if elt := c.t2.Lookup(key); elt != nil {
		c.mu.RUnlock()
		rl = true
		c.mu.Lock()
		item := c.items[key]
		if !item.IsExpired(nil) {
			c.t2.MoveToFront(elt)
			c.mu.Unlock()
			return item.value, nil
		}
		c.t2.Remove(key, elt)
		c.b2.PushFront(key)
		if c.evictedFunc != nil {
			go (*c.evictedFunc)(key, elt.Value)
		}
		c.mu.Unlock()
	}

	if !rl {
		c.mu.RUnlock()
	}

	if c.loaderFunc == nil {
		return nil, NotFoundKeyError
	}

	item, err := c.load(key, func(v interface{}, e error) (interface{}, error) {
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
	return item.(*arcItem).value, nil
}

// Remove removes the provided key from the cache.
func (c *ARC) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *ARC) remove(key interface{}) bool {
	if elt := c.t1.Lookup(key); elt != nil {
		v := elt.Value.(*arcItem).value
		c.t1.Remove(key, elt)
		if c.evictedFunc != nil {
			go (*c.evictedFunc)(key, v)
		}
		return true
	}

	if elt := c.t2.Lookup(key); elt != nil {
		v := elt.Value.(*arcItem).value
		c.t2.Remove(key, elt)
		if c.evictedFunc != nil {
			go (*c.evictedFunc)(key, v)
		}
		return true
	}

	return false
}

// Keys returns a slice of the keys in the cache.
func (c *ARC) Keys() []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := []interface{}{}
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *ARC) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Purge is used to completely clear the cache
func (c *ARC) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[interface{}]*arcItem)
	c.t1 = newARCList()
	c.t2 = newARCList()
	c.b1 = newARCList()
	c.b2 = newARCList()
}

func (c *ARC) gc() {
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

// returns boolean value whether this item is expired or not.
func (it *arcItem) IsExpired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		t := time.Now()
		now = &t
	}
	return it.expiration.Before(*now)
}

type arcList struct {
	l    *list.List
	keys map[interface{}]*list.Element
}

type arcItem struct {
	key        interface{}
	value      interface{}
	expiration *time.Time
}

func newARCList() *arcList {
	return &arcList{
		l:    list.New(),
		keys: make(map[interface{}]*list.Element),
	}
}

func (al *arcList) Has(key interface{}) bool {
	_, ok := al.keys[key]
	return ok
}

func (al *arcList) Lookup(key interface{}) *list.Element {
	elt := al.keys[key]
	return elt
}

func (al *arcList) MoveToFront(elt *list.Element) {
	al.l.MoveToFront(elt)
}

func (al *arcList) PushFront(key interface{}) {
	elt := al.l.PushFront(key)
	al.keys[key] = elt
}

func (al *arcList) Remove(key interface{}, elt *list.Element) {
	delete(al.keys, key)
	al.l.Remove(elt)
}

func (al *arcList) RemoveTail() interface{} {
	elt := al.l.Back()
	al.l.Remove(elt)

	key := elt.Value
	delete(al.keys, key)

	return key
}

func (al *arcList) Len() int {
	return al.l.Len()
}
