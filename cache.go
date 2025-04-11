package cache

import (
	"context"
	"errors"
	"time"

	"github.com/alaingilbert/clockwork"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

// DefaultCleanupInterval is exported so that someone could override the value in their project
var DefaultCleanupInterval = 10 * time.Minute

var ErrItemAlreadyExists = errors.New("item already exists")
var ErrItemNotFound = errors.New("item does not exists")

// Item wrap the user provided value and add data to it
type Item[V any] struct {
	value      V
	expiration int64
}

type Cache[K comparable, V any] struct {
	ctx               context.Context      // Context is used to stop the auto-cleanup thread
	cancel            context.CancelFunc   // Cancel the context and stop the auto-cleanup thread
	defaultExpiration time.Duration        // Default expiration for items in cache
	clock             clockwork.Clock      // Clock object for time related features
	items             RWMtxMap[K, Item[V]] // Hashmap that contains all items in the cache
}

type Config struct {
	ctx             context.Context
	cleanupInterval *time.Duration
	clock           clockwork.Clock
}

func (c *Config) WithContext(ctx context.Context) *Config {
	if ctx != nil {
		c.ctx = ctx
	}
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
	d     time.Duration
	clock clockwork.Clock
}

// Duration ...
func (c *ItemConfig) Duration(d time.Duration) *ItemConfig {
	c.d = d
	return c
}

// ItemOption ...
type ItemOption func(cfg *ItemConfig)

// NoExpire ...
func NoExpire(cfg *ItemConfig) {
	cfg = cfg.Duration(NoExpiration)
}

// ExpireIn can be used to override the default expiration for a key when calling the Add/Set/Replace methods
func ExpireIn(d time.Duration) ItemOption {
	return func(cfg *ItemConfig) {
		cfg = cfg.Duration(d)
	}
}

// ExpireAt ...
func ExpireAt(t time.Time) ItemOption {
	return func(cfg *ItemConfig) {
		cfg = cfg.Duration(cfg.clock.Until(t))
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
	c.destroy()
}

// Has returns either or not the key is present in the cache
func (c *Cache[K, V]) Has(k K) (found bool) {
	return c.has(k)
}

// Get a value associated to the given key
func (c *Cache[K, V]) Get(k K) (value V, found bool) {
	return c.get(k)
}

// GetWithExpiration a value and it's expiration
func (c *Cache[K, V]) GetWithExpiration(k K) (value V, expiration time.Time, found bool) {
	return c.getWithExpiration(k)
}

// Set a key/value pair in the cache
func (c *Cache[K, V]) Set(k K, v V, opts ...ItemOption) {
	c.set(k, v, opts...)
}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *Cache[K, V]) Add(k K, v V, opts ...ItemOption) error {
	return c.add(k, v, opts...)
}

// Replace set a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (c *Cache[K, V]) Replace(k K, v V, opts ...ItemOption) error {
	return c.replace(k, v, opts...)
}

// Delete an item from the cache
func (c *Cache[K, V]) Delete(k K) {
	c.delete(k)
}

// DeleteExpired deletes all expired items from the cache
func (c *Cache[K, V]) DeleteExpired() {
	c.deleteExpired()
}

// DeleteAll deletes all items from the cache
func (c *Cache[K, V]) DeleteAll() {
	c.deleteAll()
}

// Len returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (c *Cache[K, V]) Len() int {
	return c.len()
}

// Items copies all unexpired items in the cache into a new map and returns it.
func (c *Cache[K, V]) Items() map[K]Item[V] {
	return c.getItems()
}

func buildConfigs[C any, F ~func(*C)](opts []F) *C {
	var cfg C
	return applyOptions(&cfg, opts)
}

func applyOptions[C any, F ~func(*C)](cfg *C, opts []F) *C {
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func newCache[K comparable, V any](defaultExpiration time.Duration, opts ...Option) *Cache[K, V] {
	cfg := buildConfigs(opts)
	cfg.ctx = Or(cfg.ctx, context.Background())
	cfg.clock = Or(cfg.clock, clockwork.NewRealClock())
	cleanupInterval := Default(cfg.cleanupInterval, DefaultCleanupInterval)
	c := new(Cache[K, V])
	c.ctx, c.cancel = context.WithCancel(cfg.ctx)
	c.clock = cfg.clock
	c.defaultExpiration = defaultExpiration
	c.items = NewMap[K, Item[V]]()
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
		c.deleteExpired()
	}
}

func (c *Cache[K, V]) destroy() {
	c.cancel()
	c.deleteAll()
}

func (c *Cache[K, V]) has(k K) bool {
	_, found := c.get(k)
	return found
}

func (c *Cache[K, V]) len() int {
	return c.items.Len()
}

func (c *Cache[K, V]) getWithExpiration(k K) (V, time.Time, bool) {
	var zero V
	now := c.clock.Now().UnixNano()
	item, found := c.items.GetKey(k)
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
	cfg := &ItemConfig{clock: c.clock}
	applyOptions(cfg, opts)
	d := Or(cfg.d, c.defaultExpiration)
	e := int64(NoExpiration)
	if d != NoExpiration {
		e = c.clock.Now().Add(d).UnixNano()
	}
	c.items.SetKey(k, Item[V]{value: v, expiration: e})
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
	c.items.Clear()
}

func (c *Cache[K, V]) delete(k K) {
	c.items.DeleteKey(k)
}

func (c *Cache[K, V]) deleteExpired() {
	now := c.clock.Now().UnixNano()
	c.items.With(func(m *map[K]Item[V]) {
		for k, item := range *m {
			if item.isExpired(now) {
				delete(*m, k)
			}
		}
	})
}

func (c *Cache[K, V]) getItems() (out map[K]Item[V]) {
	now := c.clock.Now().UnixNano()
	c.items.RWith(func(m map[K]Item[V]) {
		out = make(map[K]Item[V], len(m))
		for k, v := range m {
			if !v.isExpired(now) {
				out[k] = v
			}
		}
	})
	return out
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

// Default ...
func Default[T any](v *T, d T) T {
	if v == nil {
		return d
	}
	return *v
}
