package bindgen

import (
	"slices"

	"go.bytecodealliance.org/wit"
)

// variantShape returns the type with the greatest size that is not a bool.
// If there are multiple types with the same size, it returns
// the first type that contains a pointer.
func variantShape(types []wit.Type) wit.Type {
	if len(types) == 0 {
		return nil
	}
	slices.SortStableFunc(types, func(a, b wit.Type) int {
		switch {
		case a.Size() > b.Size():
			return -1
		case a.Size() < b.Size():
			return 1
		case !isBool(a) && isBool(b):
			// bool cannot be used as variant shape
			// See https://github.com/bytecodealliance/go-modules/issues/284
			return -1
		case isBool(a) && !isBool(b):
			return 1
		case wit.HasPointer(a) && !wit.HasPointer(b):
			return -1
		case !wit.HasPointer(a) && wit.HasPointer(b):
			return 1
		default:
			return 0
		}
	})
	return types[0]
}

func isBool(t wit.TypeDefKind) bool {
	switch t := t.(type) {
	case wit.Bool:
		return true
	case *wit.TypeDef:
		return isBool(t.Root().Kind)
	}
	return false
}

// variantAlign returns the type with the largest alignment.
func variantAlign(types []wit.Type) wit.Type {
	if len(types) == 0 {
		return nil
	}
	slices.SortStableFunc(types, func(a, b wit.Type) int {
		switch {
		case a.Align() > b.Align():
			return -1
		case a.Align() < b.Align():
			return 1
		default:
			return 0
		}
	})
	return types[0]
}
