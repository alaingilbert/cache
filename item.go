package cache

import "time"

// Item wrap the user provided value and add data to it
type Item[V any] struct {
	value      V
	expiration int64
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
