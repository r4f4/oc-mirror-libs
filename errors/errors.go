// Package errors contains a common error definition for the library.
package errors

import (
	"errors"
	"fmt"
)

type ErrorKind uint

const (
	CatalogErrorKind ErrorKind = iota
)

var (
	ErrNotFound = errors.New("not found")

	// Catalog errors
	ErrCantLoad      = errors.New("cannot load catalog")
	ErrParseProperty = errors.New("cannot parse property")
)

type Error struct {
	kind   ErrorKind
	source error
}

func (e *Error) Error() string {
	var kind string
	switch e.kind {
	case CatalogErrorKind:
		kind = "catalog error"
	default:
		kind = "unknown error"
	}
	return fmt.Sprintf("%s: %s", kind, e.source)
}

func NewCatalogErr(src error) *Error {
	return &Error{kind: CatalogErrorKind, source: src}
}

func (e *Error) Is(other error) bool {
	return e == other || errors.Is(e.source, other)
}
