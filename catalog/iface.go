// Package catalog contains type definitions for working with RedHat operator catalogs.
package catalog

type CatalogIntrospector interface {
	// GetOperators returns a list of all operators for a catalog.
	GetOperators() ([]Package, error)
	// GetChannelsForOperator returns a list of all the channels for a given operator.
	GetChannelsForOperator(operatorName string) ([]Channel, error)
	// GetBundlesForChannel returns a list of bundles for a given operator's channel.
	GetBundlesForChannel(operatorName string, channelName string) ([]Bundle, error)
	// GetRelatedImagesForBundle returns a list of related images for an operator's bundle.
	GetRelatedImagesForBundle(operatorName string, bundleName string) ([]RelatedImage, error)
	// GetDependenciesForBundle returns a list of required packages for a given operator's bundle.
	GetDependenciesForBundle(operatorName string, bundleName string) ([]PackageRequired, error)
}
