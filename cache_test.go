package cache

import (
	"testing"
	"time"

	"github.com/alaingilbert/clockwork"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New[int](time.Minute, time.Minute)
	c.Set("key1", 1, DefaultExpiration)
	v, found := c.Get("key1")
	assert.True(t, found)
	assert.Equal(t, v, 1)
}

func TestNewWithKey(t *testing.T) {
	c := NewWithKey[int, int](time.Minute, time.Minute)
	c.Set(1, 1, DefaultExpiration)
	v, found := c.Get(1)
	assert.True(t, found)
	assert.Equal(t, v, 1)
	v, found = c.Get(2)
	assert.False(t, found)
	assert.Equal(t, v, 0)
}

func TestItemIsExpired(t *testing.T) {
	clock := clockwork.NewFakeClock()
	i := item[int]{value: 1, expiration: clock.Now().Add(time.Minute).UnixNano()}
	assert.False(t, i.isExpired(clock.Now().UnixNano()))
	clock.Advance(59 * time.Second)
	assert.False(t, i.isExpired(clock.Now().UnixNano()))
	clock.Advance(time.Second)
	assert.False(t, i.isExpired(clock.Now().UnixNano()))
	clock.Advance(time.Second)
	assert.True(t, i.isExpired(clock.Now().UnixNano()))
}
