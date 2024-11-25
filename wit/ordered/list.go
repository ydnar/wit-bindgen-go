package ordered

import (
	"go.bytecodealliance.org/wit/iterate"
)

type list[K, V any] struct {
	root element[K, V]
}

func (l *list[K, V]) all() iterate.Seq2[K, V] {
	return func(yield func(k K, v V) bool) {
		next := l.root.next
		for e := next; e != nil; e = next {
			// Save the next element in case e is deleted from the list.
			next = e.next
			if !yield(e.k, e.v) {
				return
			}
			// Check if an element was added to the list.
			if next == nil && e.next != nil {
				next = e.next
			}
		}
	}
}

func (l *list[K, V]) pushBack(k K, v V) *element[K, V] {
	e := &element[K, V]{k: k, v: v}
	if l.root.prev == nil {
		l.root.prev = e
		l.root.next = e
		return e
	}
	e.prev = l.root.prev
	l.root.prev.next = e
	l.root.prev = e
	return e
}

func (l *list[K, V]) delete(e *element[K, V]) {
	if e.prev == nil {
		l.root.next = e.next
	} else {
		e.prev.next = e.next
	}
	if e.next == nil {
		l.root.prev = e.prev
	} else {
		e.next.prev = e.prev
	}
	e.next = nil
	e.prev = nil
}

type element[K, V any] struct {
	prev, next *element[K, V]
	k          K
	v          V
}
