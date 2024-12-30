package cm

// IndexFunc is a function that returns an integer index of s.
// Used to reverse a string into a [variant] or [enum] case.
//
// [enum]: https://component-model.bytecodealliance.org/design/wit.html#enums
// [variant]: https://component-model.bytecodealliance.org/design/wit.html#variants
type IndexFunc func(s string) int

// Index returns an [IndexFunc] that indexes the strings slice.
func Index(strings []string) IndexFunc {
	if len(strings) <= linearScanThreshold {
		return linearIndex[string](strings).indexOf
	}
	m := make(mapIndex[string], len(strings))
	for i, v := range strings {
		m[v] = i
	}
	return m.indexOf
}

const linearScanThreshold = 16

type linearIndex[V comparable] []V

func (idx linearIndex[V]) indexOf(v V) int {
	for i := 0; i < len(idx); i++ {
		if idx[i] == v {
			return i
		}
	}
	return -1
}

type mapIndex[V comparable] map[V]int

func (idx mapIndex[V]) indexOf(v V) int {
	i, ok := idx[v]
	if !ok {
		return -1
	}
	return i
}
