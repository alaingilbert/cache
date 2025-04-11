package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/alaingilbert/clockwork"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

var DefaultCleanupInterval = 10 * time.Minute

var ErrItemAlreadyExists = errors.New("item already exists")
var ErrItemNotFound = errors.New("item does not exists")

// Item wrap the user provided value and add data to it
type Item[V any] struct {
	value      V
	expiration int64
}

type Cache[K comparable, V any] struct {
	ctx               context.Context    // Context is used to stop the auto-cleanup thread
	cancel            context.CancelFunc // Cancel the context and stop the auto-cleanup thread
	defaultExpiration time.Duration      // Default expiration for items in cache
	clock             clockwork.Clock    // Clock object for time related features
	mtx               sync.RWMutex       // This mutex should only be used in exported methods
	items             map[K]Item[V]      // Hashmap that contains all items in the cache
}

type Config struct {
	ctx             context.Context
	cleanupInterval *time.Duration
	clock           clockwork.Clock
}

func (c *Config) WithContext(ctx context.Context) *Config {
	if ctx == nil {
		panic("nil context")
	}
	c.ctx = ctx
	return c
}

func (c *Config) CleanupInterval(cleanupInterval time.Duration) *Config {
	if cleanupInterval != 0 {
		c.cleanupInterval = &cleanupInterval
	}
	return c
}

func (c *Config) WithClock(clock clockwork.Clock) *Config {
	if clock != nil {
		c.clock = clock
	}
	return c
}

type Option func(cfg *Config)

// WithContext changes context of the request.
func WithContext(ctx context.Context) Option {
	return func(cfg *Config) {
		cfg = cfg.WithContext(ctx)
	}
}

// CleanupInterval changes the cleanup interval
func CleanupInterval(cleanupInterval time.Duration) Option {
	return func(cfg *Config) {
		cfg = cfg.CleanupInterval(cleanupInterval)
	}
}

// WithClock changes the clock
func WithClock(clock clockwork.Clock) Option {
	return func(cfg *Config) {
		cfg = cfg.WithClock(clock)
	}
}

// ItemConfig ...
type ItemConfig struct {
	d time.Duration
}

// Duration ...
func (c *ItemConfig) Duration(d time.Duration) *ItemConfig {
	c.d = d
	return c
}

// ItemOption ...
type ItemOption func(cfg *ItemConfig)

// ExpireIn can be used to override the default expiration for a key when calling the Add/Set/Replace methods
func ExpireIn(d time.Duration) ItemOption {
	return func(cfg *ItemConfig) {
		cfg = cfg.Duration(d)
	}
}

// New creates a cache with K as string
func New[V any](defaultExpiration time.Duration, opts ...Option) *Cache[string, V] {
	return newCache[string, V](defaultExpiration, opts...)
}

// NewWithKey creates a cache with a custom comparable K provided by the user
func NewWithKey[K comparable, V any](defaultExpiration time.Duration, opts ...Option) *Cache[K, V] {
	return newCache[K, V](defaultExpiration, opts...)
}

// Destroy the cache object, cleanup all resources
func (c *Cache[K, V]) Destroy() {
	c.mtx.Lock()
	c.destroy()
	c.mtx.Unlock()
}

// Has returns either or not the key is present in the cache
func (c *Cache[K, V]) Has(k K) bool {
	c.mtx.RLock()
	found := c.has(k)
	c.mtx.RUnlock()
	return found
}

// Get a value associated to the given key
func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mtx.RLock()
	value, found := c.get(k)
	c.mtx.RUnlock()
	return value, found
}

// GetWithExpiration a value and it's expiration
func (c *Cache[K, V]) GetWithExpiration(k K) (V, time.Time, bool) {
	c.mtx.RLock()
	value, expiration, found := c.getWithExpiration(k)
	c.mtx.RUnlock()
	return value, expiration, found
}

// Set a key/value pair in the cache
func (c *Cache[K, V]) Set(k K, v V, opts ...ItemOption) {
	c.mtx.Lock()
	c.set(k, v, opts...)
	c.mtx.Unlock()
}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *Cache[K, V]) Add(k K, v V, opts ...ItemOption) error {
	c.mtx.Lock()
	err := c.add(k, v, opts...)
	c.mtx.Unlock()
	return err
}

// Replace set a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (c *Cache[K, V]) Replace(k K, v V, opts ...ItemOption) error {
	c.mtx.Lock()
	err := c.replace(k, v, opts...)
	c.mtx.Unlock()
	return err
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

// DeleteAll deletes all items from the cache
func (c *Cache[K, V]) DeleteAll() {
	c.mtx.Lock()
	c.deleteAll()
	c.mtx.Unlock()
}

// Len returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (c *Cache[K, V]) Len() int {
	c.mtx.Lock()
	n := len(c.items)
	c.mtx.Unlock()
	return n
}

// Items copies all unexpired items in the cache into a new map and returns it.
func (c *Cache[K, V]) Items() map[K]Item[V] {
	c.mtx.RLock()
	items := c.getItems()
	c.mtx.RUnlock()
	return items
}

func newCache[K comparable, V any](defaultExpiration time.Duration, opts ...Option) *Cache[K, V] {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.ctx = Or(cfg.ctx, context.Background())
	cfg.clock = Or(cfg.clock, clockwork.NewRealClock())
	cleanupInterval := DefaultCleanupInterval
	if cfg.cleanupInterval != nil {
		cleanupInterval = *cfg.cleanupInterval
	}
	c := new(Cache[K, V])
	c.ctx, c.cancel = context.WithCancel(cfg.ctx)
	c.clock = cfg.clock
	c.defaultExpiration = defaultExpiration
	c.items = make(map[K]Item[V])
	if cleanupInterval > 0 {
		go c.autoCleanup(cleanupInterval)
	}
	return c
}

func (c *Cache[K, V]) autoCleanup(cleanupInterval time.Duration) {
	for {
		select {
		case <-time.After(cleanupInterval):
		case <-c.ctx.Done():
			return
		}
		// Important to call the exported method to lock the mutex
		c.DeleteExpired()
	}
}

func (c *Cache[K, V]) destroy() {
	c.cancel()
	clear(c.items)
	c = nil
}

func (c *Cache[K, V]) has(k K) bool {
	_, found := c.get(k)
	return found
}

func (c *Cache[K, V]) getWithExpiration(k K) (V, time.Time, bool) {
	var zero V
	now := c.clock.Now().UnixNano()
	item, found := c.items[k]
	if !found {
		return zero, time.Time{}, false
	}
	e := time.Time{}
	if item.expiration > 0 {
		if item.expiration < now {
			return zero, time.Time{}, false
		}
		e = time.Unix(0, item.expiration)
	}
	return item.value, e, found
}

func (c *Cache[K, V]) get(k K) (V, bool) {
	value, _, found := c.getWithExpiration(k)
	return value, found
}

func (c *Cache[K, V]) set(k K, v V, opts ...ItemOption) {
	cfg := &ItemConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	e := int64(NoExpiration)
	d := cfg.d
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	e = c.clock.Now().Add(d).UnixNano()
	c.items[k] = Item[V]{value: v, expiration: e}
}

func (c *Cache[K, V]) add(k K, v V, opts ...ItemOption) error {
	if _, found := c.get(k); found {
		return ErrItemAlreadyExists
	}
	c.set(k, v, opts...)
	return nil
}

func (c *Cache[K, V]) replace(k K, v V, opts ...ItemOption) error {
	if _, found := c.get(k); !found {
		return ErrItemNotFound
	}
	c.set(k, v, opts...)
	return nil
}

func (c *Cache[K, V]) deleteAll() {
	c.items = make(map[K]Item[V])
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

func (c *Cache[K, V]) getItems() map[K]Item[V] {
	now := c.clock.Now().UnixNano()
	m := make(map[K]Item[V], len(c.items))
	for k, v := range c.items {
		if !v.isExpired(now) {
			m[k] = v
		}
	}
	return m
}

// Value returns the value contained by the item
func (i Item[V]) Value() V {
	return i.value
}

// Expiration returns the expiration time
func (i Item[V]) Expiration() time.Time {
	return time.Unix(0, i.expiration)
}

// IsExpired returns either or not the item is expired right now
func (i Item[V]) IsExpired() bool {
	now := time.Now().UnixNano()
	return i.isExpired(now)
}

// Given a unix (nano) timestamp, return either or not the item is expired
func (i Item[V]) isExpired(ts int64) bool {
	return i.expiration > 0 && i.expiration < ts
}

// Ternary ...
func Ternary[T any](predicate bool, a, b T) T {
	if predicate {
		return a
	}
	return b
}

// Or return "a" if it is non-zero otherwise "b"
func Or[T comparable](a, b T) (zero T) {
	return Ternary(a != zero, a, b)
}
