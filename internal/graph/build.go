package graph

import (
	"fmt"

	"swift-deps-diagram/internal/manifest"
)

func targetNodeID(name string) string {
	return "target::" + name
}

func productNodeID(name, pkg string) string {
	if pkg != "" {
		return fmt.Sprintf("pkg::%s::%s", pkg, name)
	}
	return "product::" + name
}

func byNameNodeID(name string) string {
	return "name::" + name
}

func addNode(nodes map[string]Node, id, label string, kind NodeKind) {
	if _, ok := nodes[id]; ok {
		return
	}
	nodes[id] = Node{ID: id, Label: label, Kind: kind}
}

func addEdge(edges *[]Edge, dedup map[string]struct{}, edge Edge) {
	key := EdgeKey(edge)
	if _, ok := dedup[key]; ok {
		return
	}
	dedup[key] = struct{}{}
	*edges = append(*edges, edge)
}

func shouldIncludeTarget(target manifest.Target, includeTests bool) bool {
	if includeTests {
		return true
	}
	return target.Type != "test"
}

// Build converts manifest targets and dependencies into a directed dependency graph.
func Build(pkg manifest.Package, includeTests bool) (Graph, error) {
	nodes := make(map[string]Node)
	edges := make([]Edge, 0)
	edgeDedup := make(map[string]struct{})

	localTargets := make(map[string]struct{})
	for _, target := range pkg.Targets {
		if !shouldIncludeTarget(target, includeTests) {
			continue
		}
		localTargets[target.Name] = struct{}{}
		addNode(nodes, targetNodeID(target.Name), target.Name, NodeKindTarget)
	}

	for _, target := range pkg.Targets {
		if !shouldIncludeTarget(target, includeTests) {
			continue
		}
		fromID := targetNodeID(target.Name)

		for _, dep := range target.Dependencies {
			switch dep.Kind {
			case manifest.DependencyKindTarget:
				if dep.Name == "" {
					continue
				}
				toID := targetNodeID(dep.Name)
				if _, ok := localTargets[dep.Name]; !ok {
					toID = productNodeID(dep.Name, "")
					addNode(nodes, toID, dep.Name, NodeKindExternalProduct)
				}
				addEdge(&edges, edgeDedup, Edge{FromID: fromID, ToID: toID, Kind: EdgeKindTarget})
			case manifest.DependencyKindProduct:
				if dep.Name == "" {
					continue
				}
				toID := productNodeID(dep.Name, dep.Package)
				addNode(nodes, toID, dep.Name, NodeKindExternalProduct)
				addEdge(&edges, edgeDedup, Edge{FromID: fromID, ToID: toID, Kind: EdgeKindProduct})
			case manifest.DependencyKindByName:
				if dep.Name == "" {
					continue
				}
				toID := byNameNodeID(dep.Name)
				if _, ok := localTargets[dep.Name]; ok {
					toID = targetNodeID(dep.Name)
				} else {
					addNode(nodes, toID, dep.Name, NodeKindExternalProduct)
				}
				addEdge(&edges, edgeDedup, Edge{FromID: fromID, ToID: toID, Kind: EdgeKindByName})
			}
		}
	}

	g := Graph{Nodes: nodes, Edges: edges}
	g.Edges = SortedEdges(g)
	return g, nil
}
