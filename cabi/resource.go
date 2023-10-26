package cabi

// TODO: remove this or move it to package cm.

// Resource is the interface implemented by all [resource] types.
type Resource[T any] interface {
	ResourceHandle() Handle[T]
	// BorrowResource() Borrow[T]
	// OwnResource() Own[T]
}

// Handle is an opaque handle to a [resource].
type Handle[T any] uint32

// Own is a handle to an owned [resource].
type Own[T any] Handle[T]

func (o Own[T]) Rep() T {
	return Rep(Handle[T](o))
}

// Borrow is a handle to a borrowed [resource].
type Borrow[T any] Handle[T]

func (b Borrow[T]) Rep() T {
	return Rep(Handle[T](b))
}

// TODO: can we use finalizers for dropping handles?

// Rep returns the representation of handle, if any.
func Rep[T any, H Handle[T]](handle H) T {
	// TODO: extract the actual representation from a table
	var v T
	return v
}
