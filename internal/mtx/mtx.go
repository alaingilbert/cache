// Package mtx provides thread-safe wrappers for values and maps using read-write mutexes.
package mtx

import "sync"

// RWMtx is a generic thread-safe wrapper for a value of type T using a RWMutex.
type RWMtx[T any] struct {
	sync.RWMutex
	v T
}

// NewRWMtx creates a new RWMtx instance with the given value.
func NewRWMtx[T any](v T) RWMtx[T] {
	return RWMtx[T]{v: v}
}

// RWith executes a read-only callback with the protected value (non-error version).
func (m *RWMtx[T]) RWith(clb func(v T)) {
	_ = m.RWithE(func(tx T) error {
		clb(tx)
		return nil
	})
}

// RWithE executes a read-only callback with the protected value (error-returning version).
func (m *RWMtx[T]) RWithE(clb func(v T) error) error {
	m.RLock()
	defer m.RUnlock()
	return clb(m.v)
}

// With executes a write callback with a pointer to the protected value (non-error version).
func (m *RWMtx[T]) With(clb func(v *T)) {
	_ = m.WithE(func(tx *T) error {
		clb(tx)
		return nil
	})
}

// WithE executes a write callback with a pointer to the protected value (error-returning version).
func (m *RWMtx[T]) WithE(clb func(v *T) error) error {
	m.Lock()
	defer m.Unlock()
	return clb(&m.v)
}

//----------------------

// RWMtxMap is a thread-safe map wrapper built on RWMtx.
type RWMtxMap[K comparable, V any] struct {
	RWMtx[map[K]V]
}

// NewRWMtxMap creates a new empty thread-safe map.
func NewRWMtxMap[K comparable, V any]() RWMtxMap[K, V] {
	return RWMtxMap[K, V]{RWMtx: NewRWMtx(make(map[K]V))}
}

// Store adds or updates a key-value pair in the map.
func (m *RWMtxMap[K, V]) Store(k K, v V) {
	m.With(func(m *map[K]V) { (*m)[k] = v })
}

// Load retrieves a value for a key and indicates existence.
func (m *RWMtxMap[K, V]) Load(k K) (out V, ok bool) {
	m.RWith(func(m map[K]V) { out, ok = m[k] })
	return
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *RWMtxMap[K, V]) LoadAndDelete(k K) (out V, loaded bool) {
	m.With(func(m *map[K]V) {
		out, loaded = (*m)[k]
		delete(*m, k)
	})
	return
}

// Delete removes a key-value pair from the map.
func (m *RWMtxMap[K, V]) Delete(k K) {
	m.With(func(m *map[K]V) { delete(*m, k) })
	return
}

// Len returns the number of elements in the map.
func (m *RWMtxMap[K, V]) Len() (out int) {
	m.RWith(func(m map[K]V) { out = len(m) })
	return
}

// Clear removes all elements from the map.
func (m *RWMtxMap[K, V]) Clear() {
	m.With(func(m *map[K]V) { clear(*m) })
}
