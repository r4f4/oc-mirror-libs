package catalog

import (
	"cmp"
	"iter"
)

type Nameable interface {
	GetName() string
}

// CompareByName compares two `nameable` instances
func CompareByName[T Nameable](a T, b T) int {
	return cmp.Compare(a.GetName(), b.GetName())
}

// EqualByName returns a comparator to check if `T` has `name`.
func EqualByName[T Nameable](name string) func(e T) bool {
	return func(e T) bool {
		return e.GetName() == name
	}
}

// CompareToName compares `T`'s name with `name`.
func CompareToName[T Nameable](e T, name string) int {
	return cmp.Compare(e.GetName(), name)
}

// Names returns an iterator for all the names in the sequence `seq`.
func Names[T Nameable](seq iter.Seq[T]) iter.Seq[string] {
	return func(yield func(string) bool) {
		for v := range seq {
			if !yield(v.GetName()) {
				return
			}
		}
	}
}
