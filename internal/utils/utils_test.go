package utils

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestPtr(t *testing.T) {
	someStr := "hello"
	someInt := 1
	assert.Equal(t, &someStr, Ptr("hello"))
	assert.NotEqual(t, &someStr, Ptr("world"))
	assert.Equal(t, &someInt, Ptr(1))
	assert.NotEqual(t, &someInt, Ptr(2))
}

func TestTernary(t *testing.T) {
	assert.Equal(t, 1, Ternary(true, 1, 2))
	assert.Equal(t, 2, Ternary(false, 1, 2))
	assert.Equal(t, "hello", Ternary(true, "hello", "world"))
	assert.Equal(t, "world", Ternary(false, "hello", "world"))
}

func TestOr(t *testing.T) {
	assert.Equal(t, "default", Or("", "default"))
	assert.Equal(t, "value", Or("value", "default"))
	assert.Equal(t, 1, Or(0, 1))
	assert.Equal(t, 2, Or(2, 1))
	assert.Equal(t, Ptr(1), Or((*int)(nil), Ptr(1)))
	assert.Equal(t, Ptr(2), Or(Ptr(2), Ptr(1)))
}

func TestDefault(t *testing.T) {
	assert.Equal(t, true, Default((*bool)(nil), true))
	assert.Equal(t, false, Default((*bool)(nil), false))
	assert.Equal(t, true, Default(Ptr(true), false))
	assert.Equal(t, false, Default(Ptr(false), true))
	assert.Equal(t, 1, Default((*int)(nil), 1))
	assert.Equal(t, 2, Default(Ptr(2), 1))
}

func TestFirst(t *testing.T) {
	assert.Equal(t, 1, First(1, 2, 3, 4))
}

func TestSecond(t *testing.T) {
	assert.Equal(t, 2, Second(1, 2, 3, 4))
}

func TestCast(t *testing.T) {
	var origin any = "test1"
	v, ok := Cast[string](origin)
	assert.True(t, ok)
	assert.Equal(t, "test1", v)
	v1, ok1 := Cast[int](origin)
	assert.False(t, ok1)
	assert.Equal(t, 0, v1)
}

func TestCastInto(t *testing.T) {
	origin := reflect.ValueOf("test1")
	var into string
	assert.True(t, CastInto(origin, &into))
	assert.Equal(t, "test1", into)

	origin1 := "test1"
	var into1 string
	assert.True(t, CastInto(origin1, &into1))
	assert.Equal(t, "test1", into1)

	origin2 := "test1"
	var into2 int
	assert.False(t, CastInto(origin2, &into2))
	assert.Equal(t, 0, into2)
}

func TestTryCast(t *testing.T) {
	var i any = Ptr(1)
	assert.True(t, TryCast[*int](i))

	var j any = reflect.ValueOf(Ptr(1))
	assert.True(t, TryCast[*int](j))
	assert.Equal(t, Ptr(1), j.(reflect.Value).Interface())
}
