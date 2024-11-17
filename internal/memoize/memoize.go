package memoize

// Function memoizes f, caching unique values of k.
// Initial calls to the resulting function will call f(k), then cache and return v.
// Subsequent calls will return the cached value for k.
func Function[F func(K) V, K comparable, V any](f F) F {
	m := make(map[K]V)
	return func(k K) V {
		if v, ok := m[k]; ok {
			return v
		}
		v := f(k)
		m[k] = v
		return v
	}
}
