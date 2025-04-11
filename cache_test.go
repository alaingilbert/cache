package cache

import (
	"testing"
	"time"

	"github.com/alaingilbert/clockwork"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New[int](time.Minute)
	c.Set("key1", 1)
	v, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, v, 1)
}

func TestNewWithKey(t *testing.T) {
	c := NewWithKey[int, int](time.Minute)
	c.Set(1, 1)
	v, found := c.Get(1)
	assert.True(t, found)
	assert.Equal(t, v, 1)
	v, found = c.Get(2)
	assert.False(t, found)
	assert.Equal(t, v, 0)
}

func TestItemIsExpired(t *testing.T) {
	clock := clockwork.NewFakeClock()
	i := Item[int]{value: 1, expiration: clock.Now().Add(time.Minute).UnixNano()}
	assert.False(t, i.isExpired(clock.Now().UnixNano()))
	clock.Advance(59 * time.Second)
	assert.False(t, i.isExpired(clock.Now().UnixNano()))
	clock.Advance(time.Second)
	assert.False(t, i.isExpired(clock.Now().UnixNano()))
	clock.Advance(time.Second)
	assert.True(t, i.isExpired(clock.Now().UnixNano()))
}

func TestExpireAt(t *testing.T) {
	clock := clockwork.NewFakeClockAt(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local))
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", ExpireAt(clock.Now().Add(15*time.Minute)))
	clock.Advance(14 * time.Minute)
	assert.True(t, c.Has("key1"))
	clock.Advance(2 * time.Minute)
	assert.False(t, c.Has("key1"))
}

func TestGetExpiredItem(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1")
	assert.True(t, c.Has("key1"))
	clock.Advance(61 * time.Second)
	assert.False(t, c.Has("key1"))
}

func TestOverrideDefaultExpiration(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", ExpireIn(5*time.Second))
	c.Set("key2", "val2")
	assert.True(t, c.Has("key1"))
	clock.Advance(4 * time.Second)
	assert.True(t, c.Has("key1"))
	clock.Advance(2 * time.Second)
	assert.False(t, c.Has("key1"))
	assert.True(t, c.Has("key2"))
}

func TestNoExpire(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", NoExpire)
	assert.True(t, c.Has("key1"))
	clock.Advance(61 * time.Second)
	assert.True(t, c.Has("key1"))
}

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestStorePointerToStruct(t *testing.T) {
	c := New[*TestStruct](time.Minute)
	c.Set("key1", &TestStruct{Num: 1})
	value1, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, 1, value1.Num)
	value1.Num++
	value2, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, 2, value2.Num)
}

func TestDeleteAll(t *testing.T) {
	c := New[string](time.Minute)
	c.Set("key1", "val1")
	c.Set("key2", "val2")
	c.Set("key3", "val3")
	assert.Equal(t, 3, c.Len())
	c.DeleteAll()
	assert.Equal(t, 0, c.Len())
}

func TestSetCache(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := NewSet[string](time.Minute, WithClock(clock))
	c.Set("key1")
	c.Set("key2")
	c.Set("key3")
	assert.Equal(t, 3, c.Len())
	assert.True(t, c.Has("key1"))
	assert.False(t, c.Has("key4"))
	c.Delete("key1")
	assert.False(t, c.Has("key1"))
	assert.True(t, c.Has("key2"))
	clock.Advance(61 * time.Second)
	assert.False(t, c.Has("key2"))
	err := c.Add("key2")
	assert.NoError(t, err)
	assert.True(t, c.Has("key2"))
	c.DeleteAll()
	assert.Equal(t, 0, c.Len())
}
