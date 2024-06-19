package mucache

import (
	"container/list"
	"demo/u"
	"sync"
)

type CacheItem struct {
	key      string
	value    string
	subtree  []string
	callArgs u.CallArgs
}

type LRUCache struct {
	capacity  int
	mu        sync.Mutex
	items     map[string]*list.Element
	evictList *list.List
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

func (c *LRUCache) Set(key string, value string, subtree []string, callArgs u.CallArgs) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setInternal(key, value, subtree, callArgs)
}

func (c *LRUCache) Get(key string) (value string, subtree []string, callArgs u.CallArgs, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getInternal(key)
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deleteInternal(key)
}

func (c *LRUCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sizeInternal()
}

func (c *LRUCache) setInternal(key string, value string, subtree []string, callArgs u.CallArgs) {
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		elem.Value.(*CacheItem).value = value
		elem.Value.(*CacheItem).subtree = subtree
		elem.Value.(*CacheItem).callArgs = callArgs
		return
	}
	if c.evictList.Len() == c.capacity {
		c.removeOldestInternal()
	}
	newItem := &CacheItem{key, value, subtree, callArgs}
	entry := c.evictList.PushFront(newItem)
	c.items[key] = entry
}

func (c *LRUCache) getInternal(key string) (string, []string, u.CallArgs, bool) {
	if elem, found := c.items[key]; found {
		c.evictList.MoveToFront(elem)
		item := elem.Value.(*CacheItem)
		return item.value, item.subtree, item.callArgs, true
	}
	return "", nil, u.CallArgs{}, false
}

func (c *LRUCache) removeOldestInternal() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElementInternal(elem)
	}
}

func (c *LRUCache) removeElementInternal(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*CacheItem)
	delete(c.items, kv.key)
}

func (c *LRUCache) deleteInternal(key string) {
	if elem, found := c.items[key]; found {
		c.removeElementInternal(elem)
	}
}

func (c *LRUCache) sizeInternal() int {
	return c.evictList.Len()
}
