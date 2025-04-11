package cache

import (
	"sync"
)

type RWMtx[T any] struct {
	sync.RWMutex
	v T
}

func NewRWMtx[T any](v T) RWMtx[T] {
	return RWMtx[T]{v: v}
}

func (m *RWMtx[T]) Val() *T {
	return &m.v
}

func (m *RWMtx[T]) Get() T {
	m.RLock()
	defer m.RUnlock()
	return m.v
}

func (m *RWMtx[T]) Set(v T) {
	m.Lock()
	defer m.Unlock()
	m.v = v
}

func (m *RWMtx[T]) Replace(newVal T) (old T) {
	m.With(func(v *T) {
		old = *v
		*v = newVal
	})
	return
}

func (m *RWMtx[T]) RWith(clb func(v T)) {
	_ = m.RWithE(func(tx T) error {
		clb(tx)
		return nil
	})
}

func (m *RWMtx[T]) RWithE(clb func(v T) error) error {
	m.RLock()
	defer m.RUnlock()
	return clb(m.v)
}

func (m *RWMtx[T]) With(clb func(v *T)) {
	_ = m.WithE(func(tx *T) error {
		clb(tx)
		return nil
	})
}

func (m *RWMtx[T]) WithE(clb func(v *T) error) error {
	m.Lock()
	defer m.Unlock()
	return clb(&m.v)
}

//----------------------

type RWMtxMap[K comparable, V any] struct {
	RWMtx[map[K]V]
}

func NewMap[K comparable, V any]() RWMtxMap[K, V] {
	return RWMtxMap[K, V]{RWMtx: NewRWMtx(make(map[K]V))}
}

func (m *RWMtxMap[K, V]) SetKey(k K, v V) {
	m.With(func(m *map[K]V) { (*m)[k] = v })
}

func (m *RWMtxMap[K, V]) GetKey(k K) (out V, ok bool) {
	m.RWith(func(m map[K]V) { out, ok = m[k] })
	return
}

func (m *RWMtxMap[K, V]) HasKey(k K) (found bool) {
	m.RWith(func(m map[K]V) { _, found = m[k] })
	return
}

func (m *RWMtxMap[K, V]) TakeKey(k K) (out V, ok bool) {
	m.With(func(m *map[K]V) {
		out, ok = (*m)[k]
		if ok {
			delete(*m, k)
		}
	})
	return
}

func (m *RWMtxMap[K, V]) DeleteKey(k K) {
	m.With(func(m *map[K]V) { delete(*m, k) })
	return
}

func (m *RWMtxMap[K, V]) Each(clb func(K, V)) {
	m.RWith(func(m map[K]V) {
		for k, v := range m {
			clb(k, v)
		}
	})
}

// Len returns the length of the slice
func (m *RWMtxMap[K, V]) Len() (out int) {
	m.RWith(func(m map[K]V) { out = len(m) })
	return
}

// Clear ...
func (m *RWMtxMap[K, V]) Clear() {
	m.With(func(m *map[K]V) { clear(*m) })
}

// Clone returns a clone of the map
func (m *RWMtxMap[K, V]) Clone() (out map[K]V) {
	m.RWith(func(mm map[K]V) {
		out = make(map[K]V, len(mm))
		for k, v := range mm {
			out[k] = v
		}
	})
	return
}
