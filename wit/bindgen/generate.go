package bindgen

import (
	"go.bytecodealliance.org/internal/go/gen"
	"go.bytecodealliance.org/wit"
)

// Go generates one or more Go packages from [wit.Resolve] res.
// It returns any error that occurs during code generation.
func Go(res *wit.Resolve, opts ...Option) ([]*gen.Package, error) {
	g, err := newGenerator(res, opts...)
	if err != nil {
		return nil, err
	}
	return g.generate()
}
