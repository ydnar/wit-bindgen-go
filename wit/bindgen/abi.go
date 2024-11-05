package bindgen

import (
	"slices"

	"go.bytecodealliance.org/wit"
)

// variantShape returns the type with the greatest size.
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
