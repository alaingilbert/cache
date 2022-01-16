package cache

import (
	"context"
	"sync"
	"time"

	"github.com/alaingilbert/clockwork"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

// Item wrap the user provided value and add data to it
type item[V any] struct {
	value      V
	expiration int64
}

type Cache[K comparable, V any] struct {
	ctx               context.Context    // Context is used to stop the auto-cleanup thread
	cancel            context.CancelFunc // Cancel the context and stop the auto-cleanup thread
	mtx               sync.RWMutex       // This mutex should only be used in exported methods
	defaultExpiration time.Duration      // Default expiration for items in cache
	items             map[K]item[V]      // Hashmap that contains all items in the cache
	clock             clockwork.Clock    // Clock object for time related features
}

// Creates a cache with K as string
func New[V any](defaultExpiration, cleanupInterval time.Duration) *Cache[string, V] {
	return newCache[string, V](context.Background(), defaultExpiration, cleanupInterval)
}

// Creates a cache with a context provided by the user
func NewWithContext[V any](ctx context.Context, defaultExpiration, cleanupInterval time.Duration) *Cache[string, V] {
	return newCache[string, V](ctx, defaultExpiration, cleanupInterval)
}

// Create a cache with a custom comparable K provided by the user
func NewWithKey[K comparable, V any](defaultExpiration, cleanupInterval time.Duration) *Cache[K, V] {
	return newCache[K, V](context.Background(), defaultExpiration, cleanupInterval)
}

// Destroy the cache object, cleanup all resources
func (c *Cache[K, V]) Destroy() {
	c.cancel()
	c = nil
}

// Has returns either or not the key is present in the cache
func (c *Cache[K, V]) Has(k K) bool {
	c.mtx.RLock()
	found := c.has(k)
	c.mtx.RUnlock()
	return found
}

// Get an value associated to the given key
func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mtx.RLock()
	value, found := c.get(k)
	c.mtx.RUnlock()
	return value, found
}

// Set a key/value pair in the cache
func (c *Cache[K, V]) Set(k K, v V, d time.Duration) {
	c.mtx.Lock()
	c.set(k, v, d)
	c.mtx.Unlock()
}

// Delete an item from the cache
func (c *Cache[K, V]) Delete(k K) {
	c.mtx.Lock()
	c.delete(k)
	c.mtx.Unlock()
}

// DeleteExpired deletes all expired items from the cache
func (c *Cache[K, V]) DeleteExpired() {
	c.mtx.Lock()
	c.deleteExpired()
	c.mtx.Unlock()
}

// SetClock set the clock object
func (c *Cache[K, V]) SetClock(clock clockwork.Clock) {
	c.clock = clock
}

func newCache[K comparable, V any](ctx context.Context, defaultExpiration, cleanupInterval time.Duration) *Cache[K, V] {
	items := make(map[K]item[V])
	c := new(Cache[K, V])
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.clock = clockwork.NewRealClock()
	c.defaultExpiration = defaultExpiration
	c.items = items
	if cleanupInterval > 0 {
		go c.autoCleanup(cleanupInterval)
	}
	return c
}

func (c *Cache[K, V]) autoCleanup(cleanupInterval time.Duration) {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(cleanupInterval):
		}
		// Important to call the exported method to lock the mutex
		c.DeleteExpired()
	}
}

func (c *Cache[K, V]) has(k K) bool {
	_, found := c.get(k)
	return found
}

func (c *Cache[K, V]) get(k K) (V, bool) {
	var zero V
	now := c.clock.Now().UnixNano()
	item, found := c.items[k]
	if !found {
		return zero, false
	}
	if item.isExpired(now) {
		return zero, false
	}
	return item.value, found
}

func (c *Cache[K, V]) set(k K, v V, d time.Duration) {
	e := int64(NoExpiration)
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	e = c.clock.Now().Add(d).UnixNano()
	c.items[k] = item[V]{value: v, expiration: e}
}

func (c *Cache[K, V]) delete(k K) {
	delete(c.items, k)
}

func (c *Cache[K, V]) deleteExpired() {
	now := c.clock.Now().UnixNano()
	for k := range c.items {
		item := c.items[k]
		if item.isExpired(now) {
			c.delete(k)
		}
	}
}

// Given a unix (nano) timestamp, return either or not the item is expired
func (i item[V]) isExpired(ts int64) bool {
	return i.expiration > 0 && i.expiration < ts
}