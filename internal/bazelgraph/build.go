package bazelgraph

import (
	"sort"
	"strings"

	"swift-deps-diagram/internal/bazel"
	"swift-deps-diagram/internal/graph"
)

func isTestRuleKind(kind string) bool {
	return strings.HasSuffix(kind, "_test")
}

func targetNodeID(label string) string {
	return "target::" + label
}

func externalNodeID(label string) string {
	return "external::" + label
}

func Build(workspace bazel.Workspace, includeTests bool) (graph.Graph, error) {
	nodes := make(map[string]graph.Node)
	edges := make([]graph.Edge, 0)
	edgeDedup := make(map[string]struct{})

	targets := append([]bazel.Target(nil), workspace.Targets...)
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Label < targets[j].Label
	})

	kindByLabel := make(map[string]string, len(targets))
	included := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		kindByLabel[target.Label] = target.Kind
		if target.Label == "" {
			continue
		}
		if !includeTests && isTestRuleKind(target.Kind) {
			continue
		}
		included[target.Label] = struct{}{}
		id := targetNodeID(target.Label)
		nodes[id] = graph.Node{ID: id, Label: target.Label, Kind: graph.NodeKindTarget}
	}

	for _, target := range targets {
		if _, ok := included[target.Label]; !ok {
			continue
		}

		fromID := targetNodeID(target.Label)
		deps := append([]string(nil), target.Deps...)
		sort.Strings(deps)
		for _, dep := range deps {
			switch {
			case strings.HasPrefix(dep, "@"):
				toID := externalNodeID(dep)
				if _, ok := nodes[toID]; !ok {
					nodes[toID] = graph.Node{ID: toID, Label: dep, Kind: graph.NodeKindExternalProduct}
				}
				edge := graph.Edge{FromID: fromID, ToID: toID, Kind: graph.EdgeKindProduct}
				key := graph.EdgeKey(edge)
				if _, ok := edgeDedup[key]; ok {
					continue
				}
				edgeDedup[key] = struct{}{}
				edges = append(edges, edge)
			case strings.HasPrefix(dep, "//"):
				if kind, ok := kindByLabel[dep]; ok && !includeTests && isTestRuleKind(kind) {
					continue
				}
				toID := targetNodeID(dep)
				if _, ok := nodes[toID]; !ok {
					nodes[toID] = graph.Node{ID: toID, Label: dep, Kind: graph.NodeKindTarget}
				}
				edge := graph.Edge{FromID: fromID, ToID: toID, Kind: graph.EdgeKindTarget}
				key := graph.EdgeKey(edge)
				if _, ok := edgeDedup[key]; ok {
					continue
				}
				edgeDedup[key] = struct{}{}
				edges = append(edges, edge)
			}
		}
	}

	g := graph.Graph{Nodes: nodes, Edges: edges}
	g.Edges = graph.SortedEdges(g)
	return g, nil
}
