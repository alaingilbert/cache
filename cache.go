package cache

import (
	"context"
	"errors"
	"github.com/alaingilbert/cache/internal/mtx"
	"github.com/alaingilbert/cache/internal/utils"
	"time"

	"github.com/alaingilbert/clockwork"
)

const (
	// NoExpiration ...
	NoExpiration time.Duration = -1
	// DefaultExpiration ...
	DefaultExpiration time.Duration = 0
)

// DefaultCleanupInterval is exported so that someone could override the value in their project
var DefaultCleanupInterval = 10 * time.Minute

// ErrItemAlreadyExists ...
var ErrItemAlreadyExists = errors.New("item already exists")

// ErrItemNotFound ...
var ErrItemNotFound = errors.New("item does not exists")

// Cache ...
type Cache[K comparable, V any] struct {
	ctx               context.Context          // Context is used to stop the auto-cleanup thread
	cancel            context.CancelFunc       // Cancel the context and stop the auto-cleanup thread
	defaultExpiration time.Duration            // Default expiration for items in cache
	clock             clockwork.Clock          // Clock object for time related features
	items             mtx.RWMtxMap[K, Item[V]] // Mutex protected hashmap that contains all items in the cache
	cleanupEvent      chan struct{}            //
}

// Config ...
type Config struct {
	ctx             context.Context
	cleanupInterval *time.Duration
	clock           clockwork.Clock
}

// WithContext ...
func (c *Config) WithContext(ctx context.Context) *Config {
	if ctx != nil {
		c.ctx = ctx
	}
	return c
}

// CleanupInterval ...
func (c *Config) CleanupInterval(cleanupInterval time.Duration) *Config {
	if cleanupInterval != 0 {
		c.cleanupInterval = &cleanupInterval
	}
	return c
}

// WithClock ...
func (c *Config) WithClock(clock clockwork.Clock) *Config {
	if clock != nil {
		c.clock = clock
	}
	return c
}

// Option ...
type Option func(cfg *Config)

// WithContext changes context
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

// GetWithExpiration gets a value and its expiration time from the cache.
// If the item never expires a zero value for time.Time is returned.
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

func newCache[K comparable, V any](defaultExpiration time.Duration, opts ...Option) *Cache[K, V] {
	cfg := utils.BuildConfig(opts)
	cfg.ctx = utils.Or(cfg.ctx, context.Background())
	cfg.clock = utils.Or(cfg.clock, clockwork.NewRealClock())
	cleanupInterval := utils.Default(cfg.cleanupInterval, DefaultCleanupInterval)
	c := new(Cache[K, V])
	c.ctx, c.cancel = context.WithCancel(cfg.ctx)
	c.clock = cfg.clock
	c.defaultExpiration = defaultExpiration
	c.items = mtx.NewRWMtxMap[K, Item[V]]()
	c.cleanupEvent = make(chan struct{})
	if cleanupInterval > 0 {
		go c.autoCleanup(cleanupInterval)
	}
	return c
}

func newSet[K comparable](defaultExpiration time.Duration, opts ...Option) *SetCache[K] {
	return &SetCache[K]{c: newCache[K, struct{}](defaultExpiration, opts...)}
}

func (c *Cache[K, V]) autoCleanup(cleanupInterval time.Duration) {
	for {
		select {
		case <-c.clock.After(cleanupInterval):
		case <-c.ctx.Done():
			return
		}
		c.deleteExpired()
		select {
		case c.cleanupEvent <- struct{}{}:
		default:
		}
	}
}

func (c *Cache[K, V]) destroy() {
	c.cancel()
	c.deleteAll()
}

func (c *Cache[K, V]) len() int {
	return c.items.Len()
}

func (c *Cache[K, V]) now() time.Time {
	return c.clock.Now()
}

func (c *Cache[K, V]) nowNano() int64 {
	return c.now().UnixNano()
}

func (c *Cache[K, V]) getWithExpiration(k K) (V, time.Time, bool) {
	var zero V
	now := c.nowNano()
	item, found := c.items.Load(k)
	if !found {
		return zero, time.Time{}, false
	}
	e := time.Time{}
	if item.expiration > 0 {
		if item.expiration < now {
			return zero, time.Time{}, false
		}
		e = item.Expiration()
	}
	return item.value, e, found
}

func (c *Cache[K, V]) get(k K) (V, bool) {
	value, _, found := c.getWithExpiration(k)
	return value, found
}

func (c *Cache[K, V]) has(k K) bool {
	return utils.Second(c.get(k))
}

func (c *Cache[K, V]) set(k K, v V, opts ...ItemOption) {
	cfg := &ItemConfig{clock: c.clock}
	utils.ApplyOptions(cfg, opts)
	d := utils.Or(cfg.d, c.defaultExpiration)
	e := int64(NoExpiration)
	if d != NoExpiration {
		e = c.now().Add(d).UnixNano()
	}
	c.items.Store(k, Item[V]{value: v, expiration: e})
}

func (c *Cache[K, V]) add(k K, v V, opts ...ItemOption) error {
	if c.has(k) {
		return ErrItemAlreadyExists
	}
	c.set(k, v, opts...)
	return nil
}

func (c *Cache[K, V]) replace(k K, v V, opts ...ItemOption) error {
	if !c.has(k) {
		return ErrItemNotFound
	}
	c.set(k, v, opts...)
	return nil
}

func (c *Cache[K, V]) deleteAll() {
	c.items.Clear()
}

func (c *Cache[K, V]) delete(k K) {
	c.items.Delete(k)
}

func (c *Cache[K, V]) deleteExpired() {
	now := c.nowNano()
	c.items.With(func(m *map[K]Item[V]) {
		for k, item := range *m {
			if item.isExpired(now) {
				delete(*m, k)
			}
		}
	})
}

func (c *Cache[K, V]) getItems() (out map[K]Item[V]) {
	now := c.nowNano()
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

// GetCast ...
func GetCast[T any, K comparable](c *Cache[K, any], k K) (value T, ok bool) {
	var zero T
	origin, found := c.get(k)
	if !found {
		return zero, false
	}
	return utils.Cast[T](origin)
}

// GetTryCast useful if you want to test if a key exists and is of a specific type
// `if GetTryCast[int]("someKey") {`
func GetTryCast[T any, K comparable](c *Cache[K, any], k K) (ok bool) {
	return utils.Second(GetCast[T, K](c, k))
}

// GetCastInto ...
func GetCastInto[T any, K comparable](c *Cache[K, any], k K, into *T) bool {
	origin, found := c.get(k)
	if !found {
		return false
	}
	return utils.CastInto[T](origin, into)
}
