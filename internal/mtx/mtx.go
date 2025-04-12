package mtx

import (
	"sync"
)

// RWMtx ...
type RWMtx[T any] struct {
	sync.RWMutex
	v T
}

// NewRWMtx ...
func NewRWMtx[T any](v T) RWMtx[T] {
	return RWMtx[T]{v: v}
}

// RWith ...
func (m *RWMtx[T]) RWith(clb func(v T)) {
	_ = m.RWithE(func(tx T) error {
		clb(tx)
		return nil
	})
}

// RWithE ...
func (m *RWMtx[T]) RWithE(clb func(v T) error) error {
	m.RLock()
	defer m.RUnlock()
	return clb(m.v)
}

// With ...
func (m *RWMtx[T]) With(clb func(v *T)) {
	_ = m.WithE(func(tx *T) error {
		clb(tx)
		return nil
	})
}

// WithE ...
func (m *RWMtx[T]) WithE(clb func(v *T) error) error {
	m.Lock()
	defer m.Unlock()
	return clb(&m.v)
}

//----------------------

// RWMtxMap ...
type RWMtxMap[K comparable, V any] struct {
	RWMtx[map[K]V]
}

// NewRWMtxMap ...
func NewRWMtxMap[K comparable, V any]() RWMtxMap[K, V] {
	return RWMtxMap[K, V]{RWMtx: NewRWMtx(make(map[K]V))}
}

// Store ...
func (m *RWMtxMap[K, V]) Store(k K, v V) {
	m.With(func(m *map[K]V) { (*m)[k] = v })
}

// Load ...
func (m *RWMtxMap[K, V]) Load(k K) (out V, ok bool) {
	m.RWith(func(m map[K]V) { out, ok = m[k] })
	return
}

// Delete ...
func (m *RWMtxMap[K, V]) Delete(k K) {
	m.With(func(m *map[K]V) { delete(*m, k) })
	return
}

// Len ...
func (m *RWMtxMap[K, V]) Len() (out int) {
	m.RWith(func(m map[K]V) { out = len(m) })
	return
}

// Clear ...
func (m *RWMtxMap[K, V]) Clear() {
	m.With(func(m *map[K]V) { clear(*m) })
}
