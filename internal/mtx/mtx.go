package mtx

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

func NewRWMtxMap[K comparable, V any]() RWMtxMap[K, V] {
	return RWMtxMap[K, V]{RWMtx: NewRWMtx(make(map[K]V))}
}

func (m *RWMtxMap[K, V]) SetKey(k K, v V) {
	m.With(func(m *map[K]V) { (*m)[k] = v })
}

func (m *RWMtxMap[K, V]) GetKey(k K) (out V, ok bool) {
	m.RWith(func(m map[K]V) { out, ok = m[k] })
	return
}

func (m *RWMtxMap[K, V]) DeleteKey(k K) {
	m.With(func(m *map[K]V) { delete(*m, k) })
	return
}

func (m *RWMtxMap[K, V]) Len() (out int) {
	m.RWith(func(m map[K]V) { out = len(m) })
	return
}

func (m *RWMtxMap[K, V]) Clear() {
	m.With(func(m *map[K]V) { clear(*m) })
}
