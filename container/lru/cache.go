package lru

import (
	"container/list"
	"sync"
)

type Cache struct {
	mu      *sync.RWMutex
	entries int
	ll      *list.List
	cache   map[interface{}]*list.Element
}

type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

func NewCache(maxEntries int) *Cache {
	return &Cache{
		mu:      &sync.RWMutex{},
		entries: maxEntries,
		ll:      list.New(),
		cache:   make(map[interface{}]*list.Element),
	}
}

func (c *Cache) cacheFetch(key Key) (*list.Element, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.cache[key]
	return v, ok
}

func (c *Cache) Add(key Key, value interface{}) {
	if ee, ok := c.cacheFetch(key); ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}

	c.mu.Lock()
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	c.mu.Unlock()

	if c.entries != 0 && c.ll.Len() > c.entries {
		c.removeOldest()
	}
}

func (c *Cache) Get(key Key) (interface{}, bool) {
	if ee, ok := c.cacheFetch(key); ok {
		c.ll.MoveToFront(ee)
		return ee.Value.(*entry).value, true
	}

	return nil, false
}

func (c *Cache) Remove(key Key) {
	if ee, ok := c.cacheFetch(key); ok {
		c.removeElement(ee)
	}
}

func (c *Cache) removeOldest() {
	ele := c.ll.Back()

	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll.Remove(e)
	kv := e.Value.(*entry)

	delete(c.cache, kv.key)
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll = list.New()
	c.cache = make(map[interface{}]*list.Element)
}
