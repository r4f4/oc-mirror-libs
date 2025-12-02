package catalog

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
	"gotest.tools/v3/assert"

	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

const validCatalog = "../testdata/catalogs/valid-catalog/"

func TestCalogClientFails(t *testing.T) {
	catalog, err := LoadCatalog(context.Background(), validCatalog)
	assert.NilError(t, err)

	t.Run("loading invalid catalog", func(t *testing.T) {
		ctlg, err := LoadCatalog(context.Background(), "/invalid/catalog/path")
		assert.ErrorIs(t, err, libErrs.ErrCantLoad)
		assert.Assert(t, ctlg == nil)
	})

	t.Run("with invalid operator", func(t *testing.T) {
		_, err := catalog.getChannel("invalid-operator", "invalid-channel")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.GetChannelsForOperator("invalid-operator")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.getBundle("invalid-operator", "invalid-bundle")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.GetBundlesForChannel("invalid-operator", "invalid-channel")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.GetRelatedImagesForBundle("invalid-operator", "invalid-bundle")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
	})

	t.Run("with invalid channel", func(t *testing.T) {
		_, err := catalog.getChannel("rhbk-operator", "invalid-channel")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.GetBundlesForChannel("rhbk-operator", "invalid-channel")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
	})

	t.Run("with invalid bundle", func(t *testing.T) {
		_, err := catalog.getBundle("rhbk-operator", "invalid-bundle")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.GetRelatedImagesForBundle("rhbk-operator", "invalid-bundle")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
		_, err = catalog.GetDependenciesForBundle("rhbk-operator", "invalid-bundle")
		assert.ErrorIs(t, err, libErrs.ErrNotFound)
	})
}

func TestCatalogClientSucceeds(t *testing.T) {
	catalog, err := LoadCatalog(context.Background(), validCatalog)
	assert.NilError(t, err)

	t.Run("getting operators", func(t *testing.T) {
		ops, err := catalog.GetOperators()
		assert.NilError(t, err)
		expected := []Package{
			Package(declcfg.Package{Schema: "olm.package", Name: "devspaces", DefaultChannel: "stable"}),
			Package(declcfg.Package{Schema: "olm.package", Name: "rhbk-operator", DefaultChannel: "stable-v26.4"}),
		}
		slices.SortFunc(ops, CompareByName)
		assert.DeepEqual(t, ops, expected)
	})

	t.Run("getting operator's channels", func(t *testing.T) {
		chs, err := catalog.GetChannelsForOperator("rhbk-operator")
		assert.NilError(t, err)
		expected := []Channel{
			Channel(
				declcfg.Channel{
					Schema:  "olm.channel",
					Name:    "stable-v26",
					Package: "rhbk-operator",
					Entries: []declcfg.ChannelEntry{
						{
							Name: "rhbk-operator.v26.0.5-opr.1",
						},
						{
							Name:     "rhbk-operator.v26.2.11-opr.1",
							Replaces: "rhbk-operator.v26.0.5-opr.1",
							Skips:    []string{"rhbk-operator.v26.0.5-opr.1"},
						},
					},
				},
			),
		}
		assert.DeepEqual(t, chs, expected)
	})

	t.Run("getting channel's bundles", func(t *testing.T) {
		bdls, err := catalog.GetBundlesForChannel("rhbk-operator", "stable-v26")
		assert.NilError(t, err)
		expected := []Bundle{
			Bundle(
				declcfg.Bundle{
					Schema:  "olm.bundle",
					Name:    "rhbk-operator.v26.0.5-opr.1",
					Package: "rhbk-operator",
					Image:   "registry.redhat.io/rhbk/keycloack-operator-bundle@sha256:deadbeef",
					Properties: []property.Property{
						{
							Type:  "olm.gvk",
							Value: json.RawMessage(`{"packageName":"rhbk-operator","version":"26.0.5-opr.1"}`),
						},
					},
					RelatedImages: []declcfg.RelatedImage{
						{
							Image: "registry.redhat.io/rhbk/keycloack-operator-bundle@sha256:deadbeef",
						},
					},
				},
			),
			Bundle(
				declcfg.Bundle{
					Schema:  "olm.bundle",
					Name:    "rhbk-operator.v26.2.11-opr.1",
					Package: "rhbk-operator",
					Image:   "registry.redhat.io/rhbk/keycloack-operator-bundle@sha256:deadbeef",
					Properties: []property.Property{
						{
							Type:  "olm.gvk",
							Value: json.RawMessage(`{"packageName":"rhbk-operator","version":"26.2.11-opr.1"}`),
						},
					},
					RelatedImages: []declcfg.RelatedImage{
						{
							Image: "registry.redhat.io/rhbk/keycloack-operator-bundle@sha256:deadbeef",
						},
					},
				},
			),
		}
		assert.DeepEqual(t, bdls, expected)
	})

	t.Run("getting bundle's related images", func(t *testing.T) {
		ri, err := catalog.GetRelatedImagesForBundle("rhbk-operator", "rhbk-operator.v26.2.11-opr.1")
		assert.NilError(t, err)
		expected := []RelatedImage{RelatedImage(declcfg.RelatedImage{Image: "registry.redhat.io/rhbk/keycloack-operator-bundle@sha256:deadbeef"})}
		assert.DeepEqual(t, ri, expected)
	})

	t.Run("getting dependencies for bundle", func(t *testing.T) {
		deps, err := catalog.GetDependenciesForBundle("devspaces", "devspacesoperator.v3.10.0")
		assert.NilError(t, err)
		expected := []PackageRequired{
			PackageRequired(property.PackageRequired{PackageName: "devworkspace-operator", VersionRange: ">=0.12.0"}),
		}
		assert.DeepEqual(t, deps, expected)
	})
}
