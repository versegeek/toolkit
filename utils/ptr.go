package utils

// Ptr returns a pointer to v. This is useful for taking the address of a
// literal or value in contexts that require a pointer, such as struct fields
// and function calls that accept *T.
//
// Example:
//
//	cfg.Timeout = utils.Ptr(30 * time.Second)
func Ptr[T any](v T) *T {
	return &v
}

// Val dereferences p and returns the value. If p is nil, it returns the zero
// value of T.
//
// Example:
//
//	name := utils.Val(cfg.Name) // "" when cfg.Name == nil
func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// PtrOr dereferences p and returns the value. If p is nil, it returns
// fallback instead.
//
// Example:
//
//	timeout := utils.PtrOr(cfg.Timeout, 30*time.Second)
func PtrOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}
