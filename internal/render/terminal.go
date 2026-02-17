package render

import (
	"sort"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
)

type terminalChild struct {
	edge graph.Edge
	node graph.Node
}

func terminalLabel(label string) string {
	return strings.ReplaceAll(label, "\n", " ")
}

func sortNodeIDsByLabelThenID(g graph.Graph, ids []string) {
	sort.Slice(ids, func(i, j int) bool {
		left := g.Nodes[ids[i]]
		right := g.Nodes[ids[j]]
		if left.Label != right.Label {
			return left.Label < right.Label
		}
		return left.ID < right.ID
	})
}

func writeTerminalChildren(
	b *strings.Builder,
	childrenByFrom map[string][]terminalChild,
	pathSeen map[string]struct{},
	nodeID string,
	prefix string,
) {
	children := childrenByFrom[nodeID]
	for i, child := range children {
		isLast := i == len(children)-1
		branch := "|-- "
		nextPrefix := prefix + "|   "
		if isLast {
			branch = "\\-- "
			nextPrefix = prefix + "    "
		}

		label := terminalLabel(child.node.Label)
		if _, ok := pathSeen[child.node.ID]; ok {
			b.WriteString("\n" + prefix + branch + label + " (*)")
			continue
		}

		b.WriteString("\n" + prefix + branch + label)
		pathSeen[child.node.ID] = struct{}{}
		writeTerminalChildren(b, childrenByFrom, pathSeen, child.node.ID, nextPrefix)
		delete(pathSeen, child.node.ID)
	}
}

// Terminal renders a dependency graph in an ASCII tree format for terminal output.
func Terminal(g graph.Graph) (string, error) {
	targetIDs := make([]string, 0)
	incomingTargetEdges := make(map[string]int)
	targetChildren := make(map[string][]string)
	for id, node := range g.Nodes {
		if node.Kind != graph.NodeKindTarget {
			continue
		}
		targetIDs = append(targetIDs, id)
		incomingTargetEdges[id] = 0
	}
	sortNodeIDsByLabelThenID(g, targetIDs)

	childrenByFrom := make(map[string][]terminalChild)
	for _, edge := range graph.SortedEdges(g) {
		fromNode, ok := g.Nodes[edge.FromID]
		if !ok {
			return "", apperrors.New(apperrors.KindRuntime, "graph edge references unknown from node", nil)
		}
		toNode, ok := g.Nodes[edge.ToID]
		if !ok {
			return "", apperrors.New(apperrors.KindRuntime, "graph edge references unknown to node", nil)
		}

		childrenByFrom[fromNode.ID] = append(childrenByFrom[fromNode.ID], terminalChild{
			edge: edge,
			node: toNode,
		})

		if fromNode.Kind == graph.NodeKindTarget && toNode.Kind == graph.NodeKindTarget {
			if _, exists := incomingTargetEdges[toNode.ID]; exists {
				incomingTargetEdges[toNode.ID]++
			}
			targetChildren[fromNode.ID] = append(targetChildren[fromNode.ID], toNode.ID)
		}
	}

	for id := range childrenByFrom {
		children := childrenByFrom[id]
		sort.Slice(children, func(i, j int) bool {
			if children[i].node.Label != children[j].node.Label {
				return children[i].node.Label < children[j].node.Label
			}
			if children[i].node.ID != children[j].node.ID {
				return children[i].node.ID < children[j].node.ID
			}
			return children[i].edge.Kind < children[j].edge.Kind
		})
		childrenByFrom[id] = children
	}

	roots := make([]string, 0)
	for _, id := range targetIDs {
		if incomingTargetEdges[id] == 0 {
			roots = append(roots, id)
		}
	}

	coveredTargets := make(map[string]struct{})
	var markCovered func(string)
	markCovered = func(id string) {
		if _, ok := coveredTargets[id]; ok {
			return
		}
		coveredTargets[id] = struct{}{}
		for _, childID := range targetChildren[id] {
			markCovered(childID)
		}
	}
	for _, rootID := range roots {
		markCovered(rootID)
	}

	for _, id := range targetIDs {
		if _, ok := coveredTargets[id]; ok {
			continue
		}
		roots = append(roots, id)
		markCovered(id)
	}

	if len(roots) == 0 {
		return "(empty)", nil
	}

	var b strings.Builder
	for i, rootID := range roots {
		if i > 0 {
			b.WriteString("\n\n")
		}

		rootNode := g.Nodes[rootID]
		b.WriteString(terminalLabel(rootNode.Label))
		pathSeen := map[string]struct{}{rootID: struct{}{}}
		writeTerminalChildren(&b, childrenByFrom, pathSeen, rootID, "")
	}
	return b.String(), nil
}
