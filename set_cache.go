package cache

import "time"

// SetCache ...
type SetCache[K comparable] struct {
	c *Cache[K, struct{}]
}

// NewSet creates a new "set" cache
func NewSet[K comparable](defaultExpiration time.Duration, opts ...Option) *SetCache[K] {
	return newSetCache[K](defaultExpiration, opts...)
}

func (s *SetCache[K]) GetExpiration(k K) (expiration time.Time, found bool) {
	_, expiration, found = s.c.getWithExpiration(k)
	return
}

func (s *SetCache[K]) Add(k K, opts ...ItemOption) error {
	return s.c.add(k, struct{}{}, opts...)
}

func (s *SetCache[K]) Set(k K, opts ...ItemOption) {
	s.c.set(k, struct{}{}, opts...)
}

func (s *SetCache[K]) Replace(k K, opts ...ItemOption) error {
	return s.c.replace(k, struct{}{}, opts...)
}

func (s *SetCache[K]) Delete(k K) {
	s.c.delete(k)
}

func (s *SetCache[K]) DeleteAll() {
	s.c.deleteAll()
}

func (s *SetCache[K]) DeleteExpired() {
	s.c.deleteExpired()
}

func (s *SetCache[K]) Has(k K) bool {
	return s.c.has(k)
}

func (s *SetCache[K]) Len() int {
	return s.c.len()
}
