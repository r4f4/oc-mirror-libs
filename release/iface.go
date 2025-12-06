// Package release contains type definitions for working with release data.
package release

import (
	"github.com/Masterminds/semver/v3"
)

type Metadata map[string]string

type Risk struct {
	Url     string `json:"url"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

type ReleaseIntrospector interface {
	GetReleases() ([]*semver.Version, error)
	GetPayload(*semver.Version) (string, error)
	GetMetadata(*semver.Version) (Metadata, error)
	GetUpdatesFrom(*semver.Version) ([]*semver.Version, error)
	GetUpdatesTo(*semver.Version) ([]*semver.Version, error)
	GetUpdatePath(from *semver.Version, to *semver.Version) ([]*semver.Version, error)
	GetUpdatePathWithRisks(from *semver.Version, to *semver.Version) ([]*semver.Version, error)
	GetRisks(from *semver.Version, to *semver.Version) ([]Risk, error)
}
