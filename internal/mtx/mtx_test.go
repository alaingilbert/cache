package mtx

import (
	"fmt"
	"sync"
	"testing"
)

func TestRWMtx(t *testing.T) {
	mtx := NewRWMtx(42)

	mtx.RWith(func(v int) {
		if v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	})

	mtx.With(func(v *int) {
		*v = 100
	})

	mtx.RWith(func(v int) {
		if v != 100 {
			t.Errorf("expected 100, got %d", v)
		}
	})
}

func TestRWMtxMap(t *testing.T) {
	m := NewRWMtxMap[string, int]()

	m.Store("a", 1)
	v, ok := m.Load("a")
	if !ok || v != 1 {
		t.Errorf("expected to load 1, got %d, ok=%v", v, ok)
	}

	m.Store("b", 2)
	m.Store("c", 3)
	if m.Len() != 3 {
		t.Errorf("expected length 3, got %d", m.Len())
	}

	m.Delete("a")
	_, ok = m.Load("a")
	if ok {
		t.Errorf("expected key 'a' to be deleted")
	}

	val, ok := m.LoadAndDelete("b")
	if !ok || val != 2 {
		t.Errorf("expected to load 2, got %d, ok=%v", v, ok)
	}
	if m.Len() != 1 {
		t.Errorf("expected length 1, got %d", m.Len())
	}

	_, ok = m.LoadAndDelete("non-existent")
	if ok {
		t.Errorf("expected to not load non-existent key")
	}

	m.Clear()
	if m.Len() != 0 {
		t.Errorf("expected map to be cleared")
	}
}

func TestRWMtxMap_ConcurrentAccess(t *testing.T) {
	m := NewRWMtxMap[int, int]()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(n * 2)

	// Writers
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			m.Store(i, i*10)
		}(i)
	}

	// Readers
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_ = m.Len()
			m.Load(i)
		}(i)
	}

	wg.Wait()
}

func TestRWMtx_WithE(t *testing.T) {
	mtx := NewRWMtx(1)

	err := mtx.WithE(func(v *int) error {
		return fmt.Errorf("fail")
	})

	if err == nil || err.Error() != "fail" {
		t.Errorf("expected error")
	}
}
