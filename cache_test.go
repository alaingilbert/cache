package cache

import (
	"context"
	"github.com/alaingilbert/cache/internal/utils"
	"github.com/alaingilbert/clockwork"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
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
	assert.Equal(t, 1, i.Value())

	i = Item[int]{value: 1, expiration: time.Now().Add(time.Minute).UnixNano()}
	assert.False(t, i.IsExpired())
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

func TestGetWithExpiration(t *testing.T) {
	clock := clockwork.NewFakeClockAt(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local))
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1")
	c.Set("key2", "val2", NoExpire)
	value, expiration, found := c.GetWithExpiration("key1")
	assert.True(t, found)
	assert.Equal(t, value, "val1")
	assert.Equal(t, clock.Now().Add(time.Minute), expiration)
	value, expiration, found = c.GetWithExpiration("key2")
	assert.True(t, found)
	assert.Equal(t, value, "val2")
	assert.True(t, expiration.IsZero())
}

func TestGetItems(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1", ExpireIn(2*time.Minute))
	c.Set("key2", "val2")
	c.Set("key3", "val3")
	clock.Advance(61 * time.Second)
	items := c.Items()
	assert.Equal(t, 1, len(items))
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

func TestAdd(t *testing.T) {
	c := New[string](time.Minute)
	err := c.Add("key1", "val1")
	assert.NoError(t, err)
	err = c.Add("key1", "val2")
	assert.ErrorIs(t, ErrItemAlreadyExists, err)
}

func TestTake(t *testing.T) {
	c := New[string](time.Minute)
	c.Set("key1", "val1")
	assert.Equal(t, 1, c.Len())
	val, ok := c.Take("key1")
	assert.True(t, ok)
	assert.Equal(t, "val1", val)
	assert.Equal(t, 0, c.Len())
}

func TestReplace(t *testing.T) {
	c := New[string](time.Minute)
	err := c.Replace("key1", "val1")
	assert.ErrorIs(t, ErrItemNotFound, err)
	c.Set("key1", "val2")
	err = c.Replace("key1", "val2")
	assert.NoError(t, err)
	assert.Equal(t, "val2", utils.First(c.Get("key1")))
}

func TestDelete(t *testing.T) {
	c := New[string](time.Minute)
	c.Set("key1", "val1")
	assert.Equal(t, 1, c.Len())
	c.Delete("key1")
	assert.Equal(t, 0, c.Len())
}

func TestDeleteExpired(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1")
	c.Set("key2", "val2")
	c.Set("key3", "val3", ExpireIn(6*time.Minute))
	assert.Equal(t, 3, c.Len())
	clock.Advance(61 * time.Second)
	c.DeleteExpired()
	assert.Equal(t, 1, c.Len())
}

func TestAutoClean(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	clock.BlockUntil(1)
	c.Set("key1", "val1")
	assert.Equal(t, 1, c.Len())
	clock.Advance(11 * time.Minute)
	<-c.cleanupEventsCh
	assert.Equal(t, 0, c.Len())
}

func TestCleanupInterval(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock), CleanupInterval(time.Hour))
	clock.BlockUntil(1)
	c.Set("key1", "val1")
	assert.Equal(t, 1, c.Len())
	clock.Advance(11 * time.Minute)
	assert.Equal(t, 1, c.Len())
	clock.Advance(50 * time.Minute)
	<-c.cleanupEventsCh
	assert.Equal(t, 0, c.Len())
}

func TestAutoCleanEarlyReturn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock), WithContext(ctx))
	c.Set("key1", "val1")
	assert.Equal(t, 1, c.Len())
	cancel()
	assert.Equal(t, 1, c.Len())
}

func TestDestroy(t *testing.T) {
	clock := clockwork.NewFakeClock()
	c := New[string](time.Minute, WithClock(clock))
	c.Set("key1", "val1")
	assert.Equal(t, 1, c.Len())
	c.Destroy()
	assert.Equal(t, 0, c.Len())
}

func TestSetCache(t *testing.T) {
	clock := clockwork.NewFakeClockAt(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local))
	c := NewSet[string](time.Minute, WithClock(clock))

	// Test basic set operations
	c.Set("key1")
	c.Set("key2")
	c.Set("key3")
	assert.Equal(t, 3, c.Len())
	assert.True(t, c.Has("key1"))
	assert.True(t, c.Has("key2"))
	assert.True(t, c.Has("key3"))
	assert.False(t, c.Has("key4"))

	// Test deletion
	c.Delete("key1")
	assert.False(t, c.Has("key1"))
	assert.Equal(t, 2, c.Len())

	// Test expiration
	clock.Advance(61 * time.Second)
	assert.False(t, c.Has("key2"))
	assert.False(t, c.Has("key3"))

	// Test Add and Replace operations
	err := c.Add("key2")
	assert.NoError(t, err)
	assert.True(t, c.Has("key2"))

	err = c.Replace("key2")
	assert.NoError(t, err)
	assert.True(t, c.Has("key2"))

	// Test expiration time
	expiration, found := c.GetExpiration("key2")
	assert.True(t, found)
	assert.Equal(t, clock.Now().Add(time.Minute), expiration)

	// Test bulk operations
	c.DeleteAll()
	assert.Equal(t, 0, c.Len())

	// Test cleanup operations
	c.Set("key1")
	c.Set("key2")
	clock.Advance(61 * time.Second)
	c.DeleteExpired()
	assert.Equal(t, 0, c.Len())

	// Test destroy
	c.Set("key1")
	c.Destroy()
	assert.Equal(t, 0, c.Len())
}

func TestGetCast(t *testing.T) {
	c := New[any](time.Minute)
	c.Set("key1", "val1")
	c.Set("key2", 1)
	c.Set("key3", reflect.ValueOf(int64(1)))
	value1, ok1 := GetCast[int](c, "key1")
	assert.Equal(t, 0, value1)
	assert.False(t, ok1)
	value2, ok2 := GetCast[string](c, "key1")
	assert.Equal(t, "val1", value2)
	assert.True(t, ok2)
	value3, ok3 := GetCast[int](c, "key2")
	assert.Equal(t, 1, value3)
	assert.True(t, ok3)
	value4, ok4 := GetCast[int64](c, "key3")
	assert.Equal(t, int64(1), value4)
	assert.True(t, ok4)
}

func TestGetTryCast(t *testing.T) {
	c1 := New[any](time.Minute)
	c1.Set("key1", "val1")
	c1.Set("key2", 1)
	assert.True(t, GetTryCast[string](c1, "key1"))
	assert.True(t, GetTryCast[int](c1, "key2"))
	assert.False(t, GetTryCast[string](c1, "key2"))
	assert.False(t, GetTryCast[string](c1, "key3"))

	c2 := NewWithKey[int, any](time.Minute)
	c2.Set(1, "val1")
	c2.Set(2, 1)
	assert.True(t, GetTryCast[string](c2, 1))
	assert.True(t, GetTryCast[int](c2, 2))
	assert.False(t, GetTryCast[string](c2, 2))
	assert.False(t, GetTryCast[string](c2, 3))
}

func TestGetCastInto(t *testing.T) {
	c1 := New[any](time.Minute)
	c1.Set("key1", "val1")
	c1.Set("key2", 1)
	c1.Set("key3", reflect.ValueOf(int64(1)))
	var v1 string
	var v2 int
	var v3 int64
	var v4 int64
	assert.True(t, GetCastInto[string](c1, "key1", &v1))
	assert.Equal(t, "val1", v1)
	assert.False(t, GetCastInto[int](c1, "key1", &v2))
	assert.Equal(t, 0, v2)
	assert.True(t, GetCastInto[int](c1, "key2", &v2))
	assert.Equal(t, 1, v2)
	assert.True(t, GetCastInto[int64](c1, "key3", &v3))
	assert.Equal(t, int64(1), v3)
	assert.False(t, GetCastInto[int64](c1, "not-exist", &v4))
	assert.Equal(t, int64(0), v4)
}
