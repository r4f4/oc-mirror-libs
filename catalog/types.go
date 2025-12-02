package catalog

import (
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
)

// Wrappers around alpha/declcfg types.
type (
	Package         declcfg.Package
	Channel         declcfg.Channel
	Bundle          declcfg.Bundle
	RelatedImage    declcfg.RelatedImage
	PackageRequired property.PackageRequired
)

// Implement the `nameable` interface

func (p Package) GetName() string {
	return p.Name
}

func (c Channel) GetName() string {
	return c.Name
}

func (b Bundle) GetName() string {
	return b.Name
}

func (r RelatedImage) GetName() string {
	return r.Name
}
