package release

import (
	"encoding/json"
	"fmt"
	"slices"

	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

type graphData struct {
	Nodes []node  `json:"nodes"`
	Edges [][]int `json:"edges"`
}

type node struct {
	Version  string            `json:"version"`
	Payload  string            `json:"payload"`
	Metadata map[string]string `json:"metadata"`
}

func parseGraphData(data []byte) (*graphData, error) {
	var gdata graphData
	if err := json.Unmarshal(data, &gdata); err != nil {
		return nil, libErrs.NewReleaseErr(fmt.Errorf("%w: %w", libErrs.ErrParseGraphData, err))
	}
	return &gdata, nil
}

func (o *graphData) findNodeIndex(ver string) (int, error) {
	idx := slices.IndexFunc(o.Nodes, func(r node) bool { return ver == r.Version })
	if idx == -1 {
		return -1, libErrs.NewReleaseErr(fmt.Errorf("%q %w", ver, libErrs.ErrNotFound))
	}
	return idx, nil
}

func (o *graphData) nodesFrom(node int) []int {
	nodes := make([]int, 0, len(o.Edges))
	for _, edge := range o.Edges {
		if edge[0] == node {
			nodes = append(nodes, edge[1])
		}
	}
	return nodes
}

func (o *graphData) nodesTo(node int) []int {
	nodes := make([]int, 0, len(o.Edges))
	for _, edge := range o.Edges {
		if edge[1] == node {
			nodes = append(nodes, edge[0])
		}
	}
	return nodes
}
