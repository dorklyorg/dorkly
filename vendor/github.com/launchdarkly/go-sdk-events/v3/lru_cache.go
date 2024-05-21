package ldevents

import (
	"container/list"
)

type lruCache struct {
	values   map[interface{}]*list.Element
	lruList  *list.List
	capacity int
}

func newLruCache(capacity int) lruCache {
	return lruCache{
		values:   make(map[interface{}]*list.Element),
		lruList:  list.New(),
		capacity: capacity,
	}
}

func (c *lruCache) clear() {
	c.values = make(map[interface{}]*list.Element)
	c.lruList.Init()
}

// Stores a value in the cache, returning true (and marking it as recently used) if it was
// already there, or false if it was newly added.
func (c *lruCache) add(value interface{}) bool {
	if c.capacity == 0 {
		return false
	}
	if e, ok := c.values[value]; ok {
		c.lruList.MoveToFront(e)
		return true
	}
	for len(c.values) >= c.capacity {
		oldest := c.lruList.Back()
		delete(c.values, oldest.Value)
		c.lruList.Remove(oldest)
	}
	e := c.lruList.PushFront(value)
	c.values[value] = e
	return false
}
