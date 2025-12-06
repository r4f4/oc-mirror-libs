package release

import (
	"fmt"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/RyanCarrier/dijkstra/v2"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

const (
	valid419GraphData = "../testdata/cincinnati/ocp-graph-data-4.19-amd64.json"
	valid420GraphData = "../testdata/cincinnati/ocp-graph-data-4.20-amd64.json"
)

func equalVersions(s []*semver.Version, t []string) cmp.Comparison {
	return func() cmp.Result {
		if len(s) != len(t) {
			return cmp.ResultFailure("slices of different lengths")
		}
		for i, v := range s {
			if !v.Equal(semver.MustParse(t[i])) {
				return cmp.ResultFailure(fmt.Sprintf("%s != %s", v.String(), t[i]))
			}
		}
		return cmp.ResultSuccess
	}
}

func oneOfVersion(s *semver.Version, t []string) cmp.Comparison {
	return func() cmp.Result {
		for _, v := range t {
			if s.Equal(semver.MustParse(v)) {
				return cmp.ResultSuccess
			}
		}
		return cmp.ResultFailure(fmt.Sprintf("%s not in target values", s.String()))
	}
}

func TestReleaseClient(t *testing.T) {
	data, err := os.ReadFile(valid419GraphData)
	assert.NilError(t, err)

	client, err := NewReleaseClient(data)
	assert.NilError(t, err)

	t.Run("should succeed when", func(t *testing.T) {
		t.Run("getting all releases", func(t *testing.T) {
			rels, err := client.GetReleases()
			assert.NilError(t, err)
			assert.Equal(t, len(rels), 43, "unexpected number of nodes")
		})

		t.Run("getting all updates from a release", func(t *testing.T) {
			rels, err := client.GetUpdatesFrom(semver.MustParse("4.19.0"))
			assert.NilError(t, err)
			assert.Equal(t, len(rels), 3, "unexpected number of updates")
			assert.Assert(t, equalVersions(rels, []string{"4.19.1", "4.19.2", "4.19.3"}), rels)
		})

		t.Run("getting updates from a blocked release", func(t *testing.T) {
			rels, err := client.GetUpdatesFrom(semver.MustParse("4.19.7"))
			assert.NilError(t, err)
			assert.Equal(t, len(rels), 0)
		})

		t.Run("getting updates to a blocked release", func(t *testing.T) {
			rels, err := client.GetUpdatesTo(semver.MustParse("4.19.7"))
			assert.NilError(t, err)
			assert.Equal(t, len(rels), 0)
		})

		t.Run("getting all updates to a release", func(t *testing.T) {
			rels, err := client.GetUpdatesTo(semver.MustParse("4.19.17"))
			assert.NilError(t, err)
			assert.Equal(t, len(rels), 4, "unexpected number of updates")
			assert.Assert(t, equalVersions(rels, []string{"4.19.13", "4.19.14", "4.19.15", "4.19.16"}))
		})

		t.Run("getting update path between two consecutive versions", func(t *testing.T) {
			rels, err := client.GetUpdatePath(semver.MustParse("4.19.0"), semver.MustParse("4.19.1"))
			assert.NilError(t, err)
			assert.Equal(t, len(rels), 2, "unexpected number of updates")
			assert.Assert(t, equalVersions(rels, []string{"4.19.0", "4.19.1"}))
		})

		t.Run("getting upgrades across channels", func(t *testing.T) {
			data2, err := os.ReadFile(valid420GraphData)
			assert.NilError(t, err)
			srcVer := semver.MustParse("4.19.13")
			tgtVer := semver.MustParse("4.20.2")
			client2, err := NewReleaseClient(data, data2)
			assert.NilError(t, err)
			path, err := client2.GetUpdatePath(srcVer, tgtVer)
			assert.NilError(t, err)
			assert.Equal(t, len(path), 4, "unexpected number of updates")
			assert.Assert(t, equalVersions(path, []string{"4.19.13", "4.19.17", "4.20.0", "4.20.2"}))
		})

		t.Run("getting upgrades across channels with conditional edges", func(t *testing.T) {
			data2, err := os.ReadFile(valid420GraphData)
			assert.NilError(t, err)
			srcVer := semver.MustParse("4.19.11")
			tgtVer := semver.MustParse("4.20.2")
			client2, err := NewReleaseClient(data, data2)
			assert.NilError(t, err)
			path, err := client2.GetUpdatePathWithRisks(srcVer, tgtVer)
			fmt.Printf("path: %+v", path)
			assert.NilError(t, err)
			assert.Equal(t, len(path), 3, "unexpected number of updates")
			assert.Assert(t, oneOfVersion(path[1], []string{"4.19.15", "4.19.16", "4.19.17"}))
			risks, err := client2.GetRisks(srcVer, path[1])
			assert.NilError(t, err)
			assert.Equal(t, len(risks), 2, "unexpected number of risks")
			risks, err = client2.GetRisks(path[1], tgtVer)
			assert.NilError(t, err)
			assert.Equal(t, len(risks), 1, "unexpected number of risks")
		})
	})

	t.Run("should fail when", func(t *testing.T) {
		invalidVer := semver.MustParse("4.21.1")

		t.Run("release version is not valid", func(t *testing.T) {
			_, err := client.GetUpdatesFrom(invalidVer)
			assert.ErrorIs(t, err, libErrs.ErrNotFound)
			_, err = client.GetUpdatesTo(invalidVer)
			assert.ErrorIs(t, err, libErrs.ErrNotFound)
			_, err = client.GetMetadata(invalidVer)
			assert.ErrorIs(t, err, libErrs.ErrNotFound)
			_, err = client.GetUpdatePath(invalidVer, semver.MustParse("4.19.13"))
			assert.ErrorIs(t, err, dijkstra.ErrVertexNotFound)
			_, err = client.GetUpdatePath(semver.MustParse("4.19.13"), invalidVer)
			assert.ErrorIs(t, err, dijkstra.ErrVertexNotFound)
		})

		t.Run("getting upgrades across channels and no path exists", func(t *testing.T) {
			data2, err := os.ReadFile(valid420GraphData)
			assert.NilError(t, err)
			srcVer := semver.MustParse("4.19.11")
			tgtVer := semver.MustParse("4.20.2")
			client2, err := NewReleaseClient(data, data2)
			assert.NilError(t, err)
			_, err = client2.GetUpdatePath(srcVer, tgtVer)
			assert.ErrorIs(t, err, libErrs.ErrNotFound)
		})
	})
}
