package cache

import (
	"sync"
	"time"
)

const (
	defaultSize int           = 1009 // Just some arbitrary prime number, first one after 1000
	defaultTTL  int64         = 3600 // 1 hour
	defaultGC   time.Duration = 10 * time.Second
)

type item struct {
	val     string
	expires int64
}

// Cache is a simple in-memory cache with expiring items.
type Cache struct {
	m map[string]item
	l sync.RWMutex
}

// New creates a new thread-safe cache, with an optional size.
// It's recommended to use a prime number when setting the size, as it will help
// filling the cache in an optimal way. See:
// https://en.wikipedia.org/wiki/Prime_number#Other_computational_applications
// TODO: verify the built in golang map will benefit with this claim.
func New(size int) *Cache {
	if size < 1 {
		size = defaultSize
	}
	return &Cache{
		m: make(map[string]item, size),
	}
}

// Len returns the number of items currently in the cache.
func (c *Cache) Len() int {
	c.l.RLock()
	l := len(c.m)
	c.l.RUnlock()
	return l
}

// Get will attempt to get a cached item using the key.
// If the item has expired, it will be automagically deleted.
func (c *Cache) Get(key string) (string, bool) {
	c.l.RLock()
	v, found := c.m[key]
	c.l.RUnlock()
	if !found {
		return "", false
	}
	if time.Now().Unix() >= v.expires {
		c.Del(key)
		return "", false
	}
	return v.val, true
}

// Set will store a new item in the cache, using the key.
// if ttl is less than a second it will use a default value.
func (c *Cache) Set(key, value string, ttl int64) {
	if ttl < 1 {
		ttl = defaultTTL
	}
	expires := time.Now().Unix() + ttl
	c.l.Lock()
	c.m[key] = item{value, expires}
	c.l.Unlock()
}

// Del will delete an item using the key.
func (c *Cache) Del(key string) {
	c.l.Lock()
	c.del(key)
	c.l.Unlock()
}

func (c *Cache) del(k string) {
	delete(c.m, k)
}

func (c *Cache) clearExpired() {
	now := time.Now().Unix()
	c.l.Lock()
	for k, v := range c.m {
		if now >= v.expires {
			c.del(k)
		}
	}
	c.l.Unlock()
}

// CancelFunc is a simple func that, when called, will stop a Cache's GC goroutine.
type CancelFunc func()

// StartGC will start a background goroutine, which will run periodically and
// clear out expired items.
// Use the returned CancelFunc to stop and remove the goroutine.
func (c *Cache) StartGC(every time.Duration) CancelFunc {
	if every < 1 {
		every = defaultGC
	}
	t := time.NewTicker(every)
	q := make(chan struct{})

	go func() {
		for {
			select {
			case <-q:
				t.Stop()
				return
			case <-t.C:
				c.clearExpired()
			}
		}
	}()
	return func() {
		close(q)
	}
}
