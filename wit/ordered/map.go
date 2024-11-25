package ordered

import (
	"go.bytecodealliance.org/internal/codec"
	"go.bytecodealliance.org/wit/clone"
	"go.bytecodealliance.org/wit/iterate"
)

// Map represents an ordered map of key-value pairs.
// Use the All method to iterate over pairs in the order they were added.
// The zero value of Map is ready to use.
// Methods on Map are not safe for concurrent use and must be protected by a synchronization mechanism.
type Map[K comparable, V any] struct {
	l list[K, V]
	m map[K]*element[K, V]
}

// Clone implements [clone.Clonable].
func (m *Map[K, V]) Clone(state *clone.State) clone.Clonable {
	var c Map[K, V]
	m.All()(func(k K, v V) bool {
		c.Set(*clone.Clone(state, &k), *clone.Clone(state, &v))
		return true
	})
	return &c
}

// Get returns a value of type V if it exists in the map, otherwise the zero value.
func (m *Map[K, V]) Get(k K) (v V) {
	if e, ok := m.m[k]; ok {
		return e.v
	}
	return
}

// GetOK returns a value of type V if it exists in the map, otherwise the zero value,
// and a boolean value that expresses whether k is present in the map.
func (m *Map[K, V]) GetOK(k K) (v V, ok bool) {
	if e, ok := m.m[k]; ok {
		return e.v, ok
	}
	return
}

// Set sets the value of k to v. If k is not present, the value is appended to the end.
// If k is already present in the map, its value is replaced.
// To guarantee the value is appended to the end, call Delete before calling Set.
// It returns true if k was present in the map and its value was replaced.
func (m *Map[K, V]) Set(k K, v V) (replaced bool) {
	if e, ok := m.m[k]; ok {
		e.v = v
		return true
	}
	e := m.l.pushBack(k, v)
	if m.m == nil {
		m.m = make(map[K]*element[K, V])
	}
	m.m[k] = e
	return
}

// Delete deletes key k from the map. It returns true if k was present in the map and deleted.
func (m *Map[K, V]) Delete(k K) (deleted bool) {
	if e, ok := m.m[k]; ok {
		delete(m.m, k)
		m.l.delete(e)
		return true
	}
	return
}

// Len returns the number of elements in m.
func (m *Map[K, V]) Len() int {
	return len(m.m)
}

// All returns a sequence that iterates over all items in m.
// It is safe to add or delete items from the map while iterating.
// New items added to the map will be yielded, deleted items will not.
func (m *Map[K, V]) All() iterate.Seq2[K, V] {
	return m.l.all()
}

// DecodeField implements the [codec.FieldDecoder] interface (if K == string).
func (m *Map[K, V]) DecodeField(dec codec.Decoder, k K) error {
	var v V
	err := dec.Decode(&v)
	if err != nil {
		return err
	}
	m.Set(k, v)
	return nil
}
