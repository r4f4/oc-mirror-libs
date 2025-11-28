package catalog

import (
	"cmp"
	"iter"
)

type nameable interface {
	GetName() string
}

// CompareByName compares two `nameable` instances
func CompareByName[T nameable](a T, b T) int {
	return cmp.Compare(a.GetName(), b.GetName())
}

// EqualByName returns a comparator to check if `T` has `name`.
func EqualByName[T nameable](name string) func(e T) bool {
	return func(e T) bool {
		return e.GetName() == name
	}
}

// CompareToName compares `T`'s name with `name`.
func CompareToName[T nameable](e T, name string) int {
	return cmp.Compare(e.GetName(), name)
}

// Names returns an iterator for all the names in the sequence `seq`.
func Names[T nameable](seq iter.Seq[T]) iter.Seq[string] {
	return func(yield func(string) bool) {
		for v := range seq {
			if !yield(v.GetName()) {
				return
			}
		}
	}
}
