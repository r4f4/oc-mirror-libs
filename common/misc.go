package common

// Map applies a function `fn` to all elements in the slice `s`.
func Map[S ~[]T, T, V any](s S, fn func(T) V) []V {
	res := make([]V, len(s))
	for i, t := range s {
		res[i] = fn(t)
	}
	return res
}
