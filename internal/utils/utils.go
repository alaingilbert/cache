package utils

import "reflect"

// Ptr ...
func Ptr[T any](v T) *T { return &v }

// Ternary ...
func Ternary[T any](predicate bool, a, b T) T {
	if predicate {
		return a
	}
	return b
}

// Or return "a" if it is non-zero otherwise "b"
func Or[T comparable](a, b T) (zero T) {
	return Ternary(a != zero, a, b)
}

// Default return the value of "v" if not nil, otherwise return the default "d" value
func Default[T any](v *T, d T) T {
	if v == nil {
		return d
	}
	return *v
}

// First returns the first argument
func First[T any](a T, _ ...any) T { return a }

// Second returns the second argument
func Second[T any](_ any, a T, _ ...any) T { return a }

// BuildConfig ...
func BuildConfig[C any, F ~func(*C)](opts []F) *C {
	var cfg C
	return ApplyOptions(&cfg, opts)
}

// ApplyOptions ...
func ApplyOptions[C any, F ~func(*C)](cfg *C, opts []F) *C {
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Cast ...
func Cast[T any](origin any) (T, bool) {
	if val, ok := origin.(reflect.Value); ok {
		origin = val.Interface()
	}
	val, ok := origin.(T)
	return val, ok
}

// TryCast ...
func TryCast[T any](origin any) bool {
	_, ok := Cast[T](origin)
	return ok
}

// CastInto ...
func CastInto[T any](origin any, into *T) bool {
	originVal, ok := origin.(reflect.Value)
	if !ok {
		originVal = reflect.ValueOf(origin)
	}
	if originVal.IsValid() {
		if _, ok := originVal.Interface().(T); ok {
			rv := reflect.ValueOf(into)
			if rv.Kind() == reflect.Pointer && !rv.IsNil() {
				rv.Elem().Set(originVal)
				return true
			}
		}
	}
	return false
}
