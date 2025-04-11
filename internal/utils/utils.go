package utils

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

// Default ...
func Default[T any](v *T, d T) T {
	if v == nil {
		return d
	}
	return *v
}

func First[T any](a T, _ ...any) T { return a }

func Second[T any](_ any, a T, _ ...any) T { return a }

func BuildConfig[C any, F ~func(*C)](opts []F) *C {
	var cfg C
	return ApplyOptions(&cfg, opts)
}

func ApplyOptions[C any, F ~func(*C)](cfg *C, opts []F) *C {
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
