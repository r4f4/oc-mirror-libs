// Package common contains shared utilities.
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// GetOCIManifest returns the manifest for the given OCI image.
func GetOCIManifest(ociPath string) (*imgspecv1.Manifest, error) {
	indexPath := filepath.Join(ociPath, "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("read oci index: %w", err)
	}

	var index imgspecv1.Index
	if err := json.Unmarshal(indexData, &index); err != nil {
		return nil, fmt.Errorf("parse oci index: %w", err)
	}

	if len(index.Manifests) == 0 {
		return nil, errors.New("no manifests found")
	}

	manifest := index.Manifests[0]
	// read the manifest blob
	manifestPath := filepath.Join(ociPath, "blobs", manifest.Digest.Algorithm().String(), manifest.Digest.Encoded())
	rawManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest blob: %w", err)
	}

	var imageManifest imgspecv1.Manifest
	if err := json.Unmarshal(rawManifest, &imageManifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	return &imageManifest, nil
}
