// Package errors contains a common error definition for the library.
package errors

import (
	"errors"
	"fmt"
)

type ErrorKind uint

const (
	CatalogErrorKind ErrorKind = iota
	ReleaseErrorKind
)

var (
	ErrNotFound = errors.New("not found")

	// Catalog errors
	ErrCantLoad      = errors.New("cannot load catalog")
	ErrParseProperty = errors.New("cannot parse property")
	ErrDownload      = errors.New("cannot download catalog")
	ErrExtract       = errors.New("cannot extract configs")

	// Release errors
	ErrParseURL       = errors.New("parse url")
	ErrParseGraphData = errors.New("cannot parse graph data")
	ErrUpdateNotFound = fmt.Errorf("update path %w", ErrNotFound)
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
	case ReleaseErrorKind:
		kind = "release error"
	default:
		kind = "unknown error"
	}
	return fmt.Sprintf("%s: %s", kind, e.source)
}

func NewCatalogErr(src error) *Error {
	return &Error{kind: CatalogErrorKind, source: src}
}

func NewReleaseErr(src error) *Error {
	return &Error{kind: ReleaseErrorKind, source: src}
}

func (e *Error) Is(other error) bool {
	return e == other || errors.Is(e.source, other)
}
