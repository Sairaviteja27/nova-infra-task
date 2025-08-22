package utils

import (
	"sync"
	"time"

	"github.com/sairaviteja27/nova-infra-task/types"
)

type cacheEntry struct {
	val    types.Result
	expiry time.Time
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
	ttl  time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]cacheEntry),
		ttl:  ttl,
	}
}

func (c *Cache) Get(addr string) (types.Result, bool) {
	c.mu.RLock()
	e, ok := c.data[addr]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiry) {
		if ok {
			c.mu.Lock()
			delete(c.data, addr)
			c.mu.Unlock()
		}
		return types.Result{}, false
	}
	return e.val, true
}

func (c *Cache) Set(addr string, val types.Result) {
	c.mu.Lock()
	c.data[addr] = cacheEntry{val: val, expiry: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
