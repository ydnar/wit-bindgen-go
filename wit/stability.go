package wit

import (
	"github.com/coreos/go-semver/semver"
	"go.bytecodealliance.org/wit/clone"
)

// Stability represents the version or feature-gated stability of a given feature.
type Stability interface {
	Node
	isStability()
}

// _stability is an embeddable type that conforms to the [Stability] interface.
type _stability struct{}

func (_stability) isStability() {}

// Stable represents a stable WIT feature, for example: @since(version = 1.2.3)
//
// Stable features have an explicit since version and an optional feature name.
type Stable struct {
	_stability
	Since      semver.Version
	Deprecated *semver.Version
}

// DeepClone implements [clone.DeepClonable].
func (s *Stable) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, s)
	c.Deprecated = clone.DeepClone(state, s.Deprecated)
	return c
}

// Unstable represents an unstable WIT feature defined by name.
type Unstable struct {
	_stability
	Feature    string
	Deprecated *semver.Version
}

// DeepClone implements [clone.DeepClonable].
func (u *Unstable) DeepClone(state *clone.State) clone.DeepClonable {
	c := clone.Clone(state, u)
	c.Deprecated = clone.DeepClone(state, u.Deprecated)
	return c
}
