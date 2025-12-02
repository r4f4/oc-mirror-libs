package release

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

const (
	OCPEndpoint string = "https://api.openshift.com/api/upgrades_info/v1/graph"
	OKDEndpoint string = "https://origin-release.ci.openshift.org/graph"
)

type Architecture string

// Supported Openshift architectures
const (
	AMD64   Architecture = "amd64"
	ARM64   Architecture = "arm64"
	S390X   Architecture = "s390x"
	PPC64LE Architecture = "ppc64le"
	MULTI   Architecture = "multi"
)

type DownloadOptions struct {
	Client   *http.Client
	Endpoint string
	Channel  string
	Arch     Architecture
}

// DownloadGraphData gets the graph data from the specified Cincinnati endpoint.
func DownloadGraphData(ctx context.Context, options DownloadOptions) ([]byte, error) {
	parsed, err := url.Parse(options.Endpoint)
	if err != nil {
		return nil, libErrs.NewReleaseErr(err)
	}
	query := parsed.Query()
	query.Set("channel", options.Channel)
	query.Set("arch", string(options.Arch))
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", parsed.String(), nil)
	if err != nil {
		return nil, libErrs.NewReleaseErr(err)
	}
	req.Header.Add("Accept", "application/json")

	client := options.Client
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, libErrs.NewReleaseErr(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if status := resp.StatusCode; status != http.StatusOK {
		return nil, libErrs.NewReleaseErr(fmt.Errorf("unexpected http status %d", status))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, libErrs.NewReleaseErr(err)
	}
	return data, nil
}
