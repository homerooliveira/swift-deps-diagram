package render

import (
	"fmt"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
)

func escapeMermaidLabel(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// Mermaid renders a dependency graph in Mermaid flowchart TD format.
func Mermaid(g graph.Graph) (string, error) {
	idList := graph.SortedNodeIDs(g)
	idMap := make(map[string]string, len(idList))
	for i, id := range idList {
		idMap[id] = fmt.Sprintf("n%d", i+1)
	}

	var b strings.Builder
	b.WriteString("flowchart TD\n")

	for _, id := range idList {
		node, ok := g.Nodes[id]
		if !ok {
			return "", apperrors.New(apperrors.KindRuntime, "graph contains missing node", nil)
		}
		b.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", idMap[id], escapeMermaidLabel(node.Label)))
	}

	for _, edge := range graph.SortedEdges(g) {
		from, okFrom := idMap[edge.FromID]
		to, okTo := idMap[edge.ToID]
		if !okFrom || !okTo {
			return "", apperrors.New(apperrors.KindRuntime, "graph edge references unknown node", nil)
		}
		b.WriteString(fmt.Sprintf("    %s --> %s\n", from, to))
	}

	return strings.TrimSpace(b.String()), nil
}
