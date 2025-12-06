package release

import (
	"encoding/json"
	"fmt"
	"slices"

	libErrs "github.com/r4f4/oc-mirror-libs/errors"
)

type graphData struct {
	Nodes     []node     `json:"nodes"`
	Edges     [][]int    `json:"edges"`
	CondEdges []condEdge `json:"conditionalEdges"`
}

type node struct {
	Version  string            `json:"version"`
	Payload  string            `json:"payload"`
	Metadata map[string]string `json:"metadata"`
}

type condEdge struct {
	Edges []edge `json:"edges"`
	Risks []Risk `json:"risks"`
}

type edge struct {
	From string `json:"from"`
	To   string `json:"to"`
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

func (o *graphData) conditionalEdgesFrom(node int) []condEdge {
	version := o.Nodes[node].Version
	condEdges := []condEdge{}
	for _, ce := range o.CondEdges {
		edges := []edge{}
		for _, e := range ce.Edges {
			if e.From != version {
				continue
			}
			edges = append(edges, e)
		}
		if len(edges) > 0 {
			condEdges = append(condEdges, condEdge{Edges: edges, Risks: ce.Risks})
		}
	}
	return condEdges
}

func (o *graphData) conditionalEdgesTo(node int) []condEdge {
	version := o.Nodes[node].Version
	condEdges := []condEdge{}
	for _, ce := range o.CondEdges {
		edges := []edge{}
		for _, e := range ce.Edges {
			if e.To != version {
				continue
			}
			edges = append(edges, e)
		}
		if len(edges) > 0 {
			condEdges = append(condEdges, condEdge{Edges: edges, Risks: ce.Risks})
		}
	}
	return condEdges
}

func (o *graphData) getRisks(from string, to string) ([]Risk, error) {
	risks := []Risk{}
	for _, ce := range o.CondEdges {
		for _, e := range ce.Edges {
			if e.From == from && e.To == to {
				risks = append(risks, ce.Risks...)
			}
		}
	}
	if len(risks) == 0 {
		return nil, libErrs.NewReleaseErr(fmt.Errorf("conditional edge %w", libErrs.ErrNotFound))
	}
	return risks, nil
}
