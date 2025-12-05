package release

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/RyanCarrier/dijkstra/v2"
	"github.com/r4f4/oc-mirror-libs/common"
	libErrs "github.com/r4f4/oc-mirror-libs/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

var logger = slog.Default().WithGroup("release")

var _ ReleaseIntrospector = (*ReleaseClient)(nil)

type ReleaseClient struct {
	data []*graphData
}

func NewReleaseClient(datas ...[]byte) (*ReleaseClient, error) {
	gdatas := make([]*graphData, len(datas))
	for i, data := range datas {
		gdata, err := parseGraphData(data)
		if err != nil {
			return nil, err
		}
		gdatas[i] = gdata
	}
	return &ReleaseClient{gdatas}, nil
}

// GetReleases returns all the release versions in a channel.
func (c *ReleaseClient) GetReleases() ([]*semver.Version, error) {
	nodes := sets.New[string]()
	for _, gdata := range c.data {
		nodes.Insert(common.Map(gdata.Nodes, func(n node) string { return n.Version })...)
	}
	rels := common.Map(nodes.UnsortedList(), semver.MustParse)
	slices.SortFunc(rels, (*semver.Version).Compare)
	return rels, nil
}

// GetPayload returns the payload for the given version.
func (c *ReleaseClient) GetPayload(ver *semver.Version) (string, error) {
	for _, gdata := range c.data {
		if idx, err := gdata.findNodeIndex(ver.String()); err == nil {
			return gdata.Nodes[idx].Payload, nil
		}
	}
	return "", libErrs.NewReleaseErr(fmt.Errorf("%q %w", ver.String(), libErrs.ErrNotFound))
}

// GetMetadata returns release metadata for the given version.
func (c *ReleaseClient) GetMetadata(ver *semver.Version) (Metadata, error) {
	for _, gdata := range c.data {
		if idx, err := gdata.findNodeIndex(ver.String()); err == nil {
			return gdata.Nodes[idx].Metadata, nil
		}
	}
	return nil, libErrs.NewReleaseErr(fmt.Errorf("%q %w", ver.String(), libErrs.ErrNotFound))
}

// GetUpdatesFrom returns the direct updates `from` version.
func (c *ReleaseClient) GetUpdatesFrom(from *semver.Version) ([]*semver.Version, error) {
	nodes := sets.New[string]()
	errs := make([]error, 0, len(c.data))
	for _, gdata := range c.data {
		idx, err := gdata.findNodeIndex(from.String())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		nodes.Insert(common.Map(gdata.nodesFrom(idx), func(i int) string { return gdata.Nodes[i].Version })...)
	}
	// Node not found in any of the graph datas
	if len(errs) == len(c.data) {
		return nil, errs[0]
	}
	semvers := common.Map(nodes.UnsortedList(), semver.MustParse)
	slices.SortFunc(semvers, (*semver.Version).Compare)
	return semvers, nil
}

// GetUpdatesTo returns the direct updates `to` version.
func (c *ReleaseClient) GetUpdatesTo(to *semver.Version) ([]*semver.Version, error) {
	nodes := sets.New[string]()
	errs := make([]error, 0, len(c.data))
	for _, gdata := range c.data {
		idx, err := gdata.findNodeIndex(to.String())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		nodes.Insert(common.Map(gdata.nodesTo(idx), func(i int) string { return gdata.Nodes[i].Version })...)
	}
	// Node not found in any of the graph datas
	if len(errs) == len(c.data) {
		return nil, errs[0]
	}
	semvers := common.Map(nodes.UnsortedList(), semver.MustParse)
	slices.SortFunc(semvers, (*semver.Version).Compare)
	return semvers, nil
}

func (c *ReleaseClient) buildGraph() (*dijkstra.MappedGraph[string], error) {
	graph := dijkstra.NewMappedGraph[string]()
	for _, gdata := range c.data {
		for _, node := range gdata.Nodes {
			if err := graph.AddEmptyVertex(node.Version); err != nil {
				logger.Debug("node already in graph", slog.String("node", node.Version))
				if !errors.Is(err, dijkstra.ErrVertexAlreadyExists) {
					return nil, err
				}
			}
		}
		for _, edge := range gdata.Edges {
			if err := graph.AddArc(gdata.Nodes[edge[0]].Version, gdata.Nodes[edge[1]].Version, 1); err != nil {
				return nil, err
			}
		}
	}
	return &graph, nil
}

// GetUpdatePath returns the update path between two releases in the same channel.
func (c *ReleaseClient) GetUpdatePath(from *semver.Version, to *semver.Version) ([]*semver.Version, error) {
	graph, err := c.buildGraph()
	if err != nil {
		return nil, libErrs.NewReleaseErr(err)
	}
	path, err := graph.Shortest(from.String(), to.String())
	if err != nil {
		if errors.Is(err, dijkstra.ErrNoPath) {
			return nil, libErrs.NewReleaseErr(libErrs.ErrUpdateNotFound)
		}
		return nil, libErrs.NewReleaseErr(err)
	}
	return common.Map(path.Path, semver.MustParse), nil
}
