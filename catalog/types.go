package catalog

import (
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
)

// Wrappers around alpha/declcfg types.

type Package struct {
	declcfg.Package
}

type Channel struct {
	declcfg.Channel
}

type Bundle struct {
	declcfg.Bundle
}

type RelatedImage struct {
	declcfg.RelatedImage
}

type PackageRequired struct {
	property.PackageRequired
}
