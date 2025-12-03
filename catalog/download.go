package catalog

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.podman.io/image/v5/copy"
	"go.podman.io/image/v5/docker"
	"go.podman.io/image/v5/docker/reference"
	"go.podman.io/image/v5/image"
	"go.podman.io/image/v5/manifest"
	"go.podman.io/image/v5/oci/layout"
	"go.podman.io/image/v5/signature"
	"go.podman.io/image/v5/types"

	"github.com/r4f4/oc-mirror-libs/common"
	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

// DownloadOptions is used to configure parameters for image download.
type DownloadOptions struct {
	DestDir        string
	ForceDownload  bool
	SystemCtx      *types.SystemContext
	Policy         *signature.Policy
	ImageSelection copy.ImageListSelection
}

// DownloadResult contains the image download output result.
type DownloadResult struct {
	Path   string
	Digest digest.Digest
}

// DownloadImageIndex downloads the given image to `destDir` in OCI format.
// The image is saved as `destDir/name/[tag]/digest/`, where `name` is `imageRef` without tag/digest.
func DownloadImageIndex(ctx context.Context, imageRef string, opts DownloadOptions) (*DownloadResult, error) {
	imageRef = strings.TrimPrefix(imageRef, "docker://")
	ref, err := docker.ParseReference("//" + imageRef)
	if err != nil {
		return nil, newDownloadErr(err)
	}

	if opts.SystemCtx == nil {
		// NOTE: catalog content is architecture-independent.
		opts.SystemCtx = &types.SystemContext{OSChoice: "linux"}
	}

	src, err := ref.NewImageSource(ctx, opts.SystemCtx)
	if err != nil {
		return nil, newDownloadErr(err)
	}
	defer func() {
		_ = src.Close()
	}()

	unparsed := image.UnparsedInstance(src, nil)
	rawManifest, _, err := unparsed.Manifest(ctx)
	if err != nil {
		return nil, newDownloadErr(err)
	}

	origDigest, err := manifest.Digest(rawManifest)
	if err != nil {
		return nil, newDownloadErr(err)
	}

	parts := reference.ReferenceRegexp.FindStringSubmatch(imageRef)
	// add tag if it's present
	ociPath := filepath.Join(opts.DestDir, parts[1], parts[2], origDigest.Encoded())

	if !opts.ForceDownload {
		// Already downloaded, nothing to do.
		if info, err := os.Stat(ociPath); err == nil && info.IsDir() {
			return &DownloadResult{Path: ociPath, Digest: origDigest}, nil
		}
	}

	if err := os.MkdirAll(ociPath, 0o755); err != nil {
		return nil, newDownloadErr(err)
	}

	destRef, err := layout.ParseReference(ociPath)
	if err != nil {
		return nil, newDownloadErr(err)
	}

	if opts.Policy == nil {
		opts.Policy, err = signature.DefaultPolicy(nil)
		if err != nil {
			return nil, newDownloadErr(err)
		}
	}
	policyCtx, err := signature.NewPolicyContext(opts.Policy)
	if err != nil {
		return nil, newDownloadErr(err)
	}
	defer func() {
		_ = policyCtx.Destroy()
	}()

	if _, err = copy.Image(
		ctx,
		policyCtx,
		destRef,
		ref,
		&copy.Options{
			SourceCtx:          opts.SystemCtx,
			DestinationCtx:     opts.SystemCtx,
			RemoveSignatures:   true, // OCI doesn't support signatures
			ImageListSelection: opts.ImageSelection,
		},
	); err != nil {
		return nil, newDownloadErr(err)
	}

	return &DownloadResult{
		Digest: origDigest,
		Path:   ociPath,
	}, nil
}

// ExtractConfigs extracts the `configs/` layer to destDir.
func ExtractConfigs(ociPath string, destDir string) error {
	imageManifest, err := common.GetOCIManifest(ociPath)
	if err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}

	for _, layer := range imageManifest.Layers {
		if err := extractLayer(ociPath, destDir, layer); err != nil {
			return err
		}
	}

	return nil
}

func extractLayer(ociPath, destPath string, layer imgspecv1.Descriptor) error {
	// Read layer blob
	layerPath := filepath.Join(ociPath, "blobs", layer.Digest.Algorithm().String(), layer.Digest.Encoded())
	layerFile, err := os.Open(layerPath)
	if err != nil {
		return newExtractErr(fmt.Errorf("open layer blob: %w", err))
	}
	defer func() {
		_ = layerFile.Close()
	}()

	var reader io.Reader = layerFile
	// decompress if gzip
	if layer.MediaType == imgspecv1.MediaTypeImageLayerGzip {
		gzReader, err := gzip.NewReader(layerFile)
		if err != nil {
			return newExtractErr(fmt.Errorf("decompress layer: %w", err))
		}
		defer func() {
			_ = gzReader.Close()
		}()
		reader = gzReader
	}

	// Extract tar archive (only configs/ directory)
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return newExtractErr(fmt.Errorf("read tar header: %w", err))
		}

		if !strings.HasPrefix(header.Name, "configs/") {
			continue
		}

		targetPath := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return newExtractErr(fmt.Errorf("create dir %s: %w", targetPath, err))
			}
		case tar.TypeReg:
			targetDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return newExtractErr(fmt.Errorf("create dir %s: %w", targetDir, err))
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return newExtractErr(fmt.Errorf("create file %s: %w", targetPath, err))
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return newExtractErr(fmt.Errorf("write file %s: %w", targetPath, err))
			}
			_ = outFile.Close()
		}
	}
	return nil
}

func newDownloadErr(err error) *libErrs.Error {
	return libErrs.NewCatalogErr(fmt.Errorf("%w: %w", libErrs.ErrDownload, err))
}

func newExtractErr(err error) *libErrs.Error {
	return libErrs.NewCatalogErr(fmt.Errorf("%w: %w", libErrs.ErrExtract, err))
}
