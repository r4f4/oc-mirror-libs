package catalog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDownload(t *testing.T) {
	t.Run("should download and extract image by tag", func(t *testing.T) {
		t.Skip("too expensive")

		const catalog string = "registry.redhat.io/redhat/redhat-operator-index:v4.19"
		tmpDir := t.TempDir()
		destDir := filepath.Join(tmpDir, "catalogs", "oci", "by-tag")
		res, err := DownloadImageIndex(context.Background(), catalog, DownloadOptions{DestDir: destDir})
		assert.NilError(t, err)
		_, err = os.Stat(res.Path)
		assert.NilError(t, err)
		destDir = filepath.Join(tmpDir, "catalogs", "extracted", "by-tag")
		err = ExtractConfigs(res.Path, destDir)
		assert.NilError(t, err)
		info, err := os.Stat(filepath.Join(destDir, "configs"))
		assert.NilError(t, err)
		assert.Assert(t, info.IsDir())
	})
	t.Run("should download and extract image by digest", func(t *testing.T) {
		t.Skip("too expensive")

		const catalog string = "registry.redhat.io/redhat/redhat-operator-index@sha256:658fba796baf221e0469b52d7982e08adad8165ebd068ab1aeba1e94c17dba6e"
		tmpdir := t.TempDir()
		destDir := filepath.Join(tmpdir, "catalogs", "oci", "by-digest")
		res, err := DownloadImageIndex(context.Background(), catalog, DownloadOptions{DestDir: destDir})
		assert.NilError(t, err)
		_, err = os.Stat(res.Path)
		assert.NilError(t, err)
		destDir = filepath.Join(tmpdir, "catalogs", "extracted", "by-digest")
		err = ExtractConfigs(res.Path, destDir)
		assert.NilError(t, err)
		info, err := os.Stat(filepath.Join(destDir, "configs"))
		assert.NilError(t, err)
		assert.Assert(t, info.IsDir())
	})
}
