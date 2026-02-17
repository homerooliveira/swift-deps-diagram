package render

import (
	"fmt"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
)

func quoteDOT(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	return "\"" + s + "\""
}

func dotStyle(kind graph.NodeKind) string {
	switch kind {
	case graph.NodeKindTarget:
		return "shape=box"
	case graph.NodeKindExternalProduct:
		return "shape=ellipse,style=dashed"
	default:
		return "shape=box"
	}
}

// Dot renders a dependency graph in Graphviz DOT format.
func Dot(g graph.Graph) (string, error) {
	var b strings.Builder
	b.WriteString("digraph dependencies {\n")
	b.WriteString("  rankdir=LR;\n")

	for _, id := range graph.SortedNodeIDs(g) {
		node, ok := g.Nodes[id]
		if !ok {
			return "", apperrors.New(apperrors.KindRuntime, "graph contains missing node", nil)
		}
		b.WriteString(fmt.Sprintf("  %s [label=%s,%s];\n", quoteDOT(node.ID), quoteDOT(node.Label), dotStyle(node.Kind)))
	}

	for _, edge := range graph.SortedEdges(g) {
		if _, ok := g.Nodes[edge.FromID]; !ok {
			return "", apperrors.New(apperrors.KindRuntime, "graph edge references unknown from node", nil)
		}
		if _, ok := g.Nodes[edge.ToID]; !ok {
			return "", apperrors.New(apperrors.KindRuntime, "graph edge references unknown to node", nil)
		}
		b.WriteString(fmt.Sprintf("  %s -> %s;\n", quoteDOT(edge.FromID), quoteDOT(edge.ToID)))
	}

	b.WriteString("}\n")
	return strings.TrimSpace(b.String()), nil
}
