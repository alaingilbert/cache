package utils

import (
	"github.com/stretchr/testify/assert"
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
