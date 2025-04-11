package cache

import (
	"testing"
	"time"

	"github.com/alaingilbert/clockwork"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New[int](time.Minute)
	c.Set("key1", 1, ExpireIn(DefaultExpiration))
	v, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, v, 1)
}

func TestNewWithKey(t *testing.T) {
	c := NewWithKey[int, int](time.Minute)
	c.Set(1, 1, ExpireIn(DefaultExpiration))
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

func TestGetExpiredItem(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", ExpireIn(DefaultExpiration))
	_, found := c.Get("key1")
	assert.True(t, found)
	clock.Advance(61 * time.Second)
	_, found = c.Get("key1")
	assert.False(t, found)
}

func TestOverrideDefaultExpiration(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", ExpireIn(5*time.Second))
	c.Set("key2", "val2")
	_, found := c.Get("key1")
	assert.True(t, found)
	clock.Advance(4 * time.Second)
	_, found = c.Get("key1")
	assert.True(t, found)
	clock.Advance(2 * time.Second)
	_, found = c.Get("key1")
	assert.False(t, found)
	_, found = c.Get("key2")
	assert.True(t, found)
}

func TestNoExpire(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", NoExpire)
	_, found := c.Get("key1")
	assert.True(t, found)
	clock.Advance(61 * time.Second)
	_, found = c.Get("key1")
	assert.True(t, found)
}

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestStorePointerToStruct(t *testing.T) {
	c := New[*TestStruct](time.Minute)
	c.Set("key1", &TestStruct{Num: 1}, ExpireIn(DefaultExpiration))
	value1, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, 1, value1.Num)
	value1.Num++
	value2, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, 2, value2.Num)
}
