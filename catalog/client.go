package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/r4f4/oc-mirror-libs/common"
	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

var logger = slog.Default().WithGroup("catalog")

var _ CatalogIntrospector = (*LoadedCatalog)(nil)

type LoadedCatalog struct {
	path string
	cfg  *declcfg.DeclarativeConfig
}

// LoadCatalog loads an OCI catalog image from `path`.
func LoadCatalog(ctx context.Context, path string) (*LoadedCatalog, error) {
	cfg, err := declcfg.LoadFS(ctx, os.DirFS(path))
	if err != nil {
		return nil, libErrs.NewCatalogErr(fmt.Errorf("%w %q: %w", libErrs.ErrCantLoad, path, err))
	}
	return &LoadedCatalog{path, cfg}, nil
}

func (l LoadedCatalog) hasOperator(name string) bool {
	return slices.ContainsFunc(l.cfg.Packages, func(e declcfg.Package) bool {
		return e.Name == name
	})
}

func (l LoadedCatalog) getChannel(operator, name string) (declcfg.Channel, error) {
	if !l.hasOperator(operator) {
		return declcfg.Channel{}, libErrs.NewCatalogErr(fmt.Errorf("operator %q %w", operator, libErrs.ErrNotFound))
	}
	idx := slices.IndexFunc(l.cfg.Channels, func(e declcfg.Channel) bool {
		return e.Package == operator && e.Name == name
	})
	if idx == -1 {
		return declcfg.Channel{}, libErrs.NewCatalogErr(fmt.Errorf("channel %q %w", name, libErrs.ErrNotFound))
	}
	return l.cfg.Channels[idx], nil
}

func (l LoadedCatalog) getBundle(operator, name string) (declcfg.Bundle, error) {
	if !l.hasOperator(operator) {
		return declcfg.Bundle{}, libErrs.NewCatalogErr(fmt.Errorf("operator %q %w", operator, libErrs.ErrNotFound))
	}
	idx := slices.IndexFunc(l.cfg.Bundles, func(e declcfg.Bundle) bool {
		return e.Package == operator && e.Name == name
	})
	if idx == -1 {
		return declcfg.Bundle{}, libErrs.NewCatalogErr(fmt.Errorf("bundle %q %w", name, libErrs.ErrNotFound))
	}
	return l.cfg.Bundles[idx], nil
}

// GetBundlesForChannel implements CatalogIntrospector.
func (l *LoadedCatalog) GetBundlesForChannel(operatorName string, channelName string) ([]Bundle, error) {
	ch, err := l.getChannel(operatorName, channelName)
	if err != nil {
		return nil, err
	}
	bundleNames := sets.New[string]()
	for _, entry := range ch.Entries {
		bundleNames.Insert(entry.Name)
	}
	lg := logger.With(slog.String("channel", channelName))
	lg.Debug("get bundles", slog.Int("unique names", bundleNames.Len()))
	foundBundles := sets.New[string]()
	bundles := make([]Bundle, 0, bundleNames.Len())
	for _, bdl := range l.cfg.Bundles {
		if bdl.Package != operatorName {
			continue
		}
		if bundleNames.Has(bdl.Name) {
			bundles = append(bundles, Bundle(bdl))
			foundBundles.Insert(bdl.Name)
		}
	}
	if diff := bundleNames.Len() - foundBundles.Len(); diff > 0 {
		lg.Warn("get bundles", slog.Int("missing", diff))
		lg.Debug("get bundles", slog.Any("not found", bundleNames.UnsortedList()))
	}
	return bundles, nil
}

// GetChannelsForOperator implements CatalogIntrospector.
func (l *LoadedCatalog) GetChannelsForOperator(operatorName string) ([]Channel, error) {
	if !l.hasOperator(operatorName) {
		return nil, libErrs.NewCatalogErr(fmt.Errorf("operator %q %w", operatorName, libErrs.ErrNotFound))
	}
	channels := []Channel{}
	for _, ch := range l.cfg.Channels {
		if ch.Package == operatorName {
			channels = append(channels, Channel(ch))
		}
	}
	return channels, nil
}

// GetDependenciesForBundle implements CatalogIntrospector.
func (l *LoadedCatalog) GetDependenciesForBundle(operatorName string, bundleName string) ([]PackageRequired, error) {
	bdl, err := l.getBundle(operatorName, bundleName)
	if err != nil {
		return nil, err
	}
	deps := make([]PackageRequired, 0, len(bdl.Properties))
	for _, prop := range bdl.Properties {
		if prop.Type != property.TypePackageRequired {
			logger.Debug("get dependencies", slog.String("skip property", prop.Type))
			continue
		}
		var v property.PackageRequired
		if err := json.Unmarshal(prop.Value, &v); err != nil {
			return nil, libErrs.NewCatalogErr(libErrs.ErrParseProperty)
		}
		deps = append(deps, PackageRequired(v))
	}
	return deps, nil
}

// GetOperators implements CatalogIntrospector.
func (l *LoadedCatalog) GetOperators() ([]Package, error) {
	return common.Map(l.cfg.Packages, func(v declcfg.Package) Package { return Package(v) }), nil
}

// GetRelatedImagesForBundle implements CatalogIntrospector.
func (l *LoadedCatalog) GetRelatedImagesForBundle(operatorName string, bundleName string) ([]RelatedImage, error) {
	bdl, err := l.getBundle(operatorName, bundleName)
	if err != nil {
		return nil, err
	}
	return common.Map(bdl.RelatedImages, func(v declcfg.RelatedImage) RelatedImage { return RelatedImage(v) }), nil
}
