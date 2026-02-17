package xcodegraph

import (
	"fmt"
	"sort"
	"strings"

	"swift-deps-diagram/internal/graph"
	"swift-deps-diagram/internal/xcodeproj"
)

func isTestTarget(productType string) bool {
	return strings.Contains(productType, "unit-test") || strings.Contains(productType, "ui-testing")
}

func targetNodeID(name, id string, seen map[string]struct{}) string {
	base := "target::" + name
	if _, ok := seen[base]; !ok {
		return base
	}
	return fmt.Sprintf("%s::%s", base, id)
}

func productNodeID(pkgIdentity, productName string) string {
	if pkgIdentity != "" {
		return fmt.Sprintf("pkg::%s::%s", pkgIdentity, productName)
	}
	return "product::" + productName
}

// Build converts parsed xcode target/product dependencies into the common graph model.
func Build(project xcodeproj.Project, includeTests bool) (graph.Graph, error) {
	nodes := make(map[string]graph.Node)
	edges := make([]graph.Edge, 0)
	edgeDedup := make(map[string]struct{})

	targetIDToNodeID := make(map[string]string)
	nodeIDSeen := make(map[string]struct{})

	targets := make([]xcodeproj.Target, 0, len(project.Targets))
	targets = append(targets, project.Targets...)
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Name != targets[j].Name {
			return targets[i].Name < targets[j].Name
		}
		return targets[i].ID < targets[j].ID
	})

	for _, target := range targets {
		if target.Name == "" {
			continue
		}
		if !includeTests && isTestTarget(target.ProductType) {
			continue
		}
		nodeID := targetNodeID(target.Name, target.ID, nodeIDSeen)
		nodeIDSeen[nodeID] = struct{}{}
		targetIDToNodeID[target.ID] = nodeID
		nodes[nodeID] = graph.Node{ID: nodeID, Label: target.Name, Kind: graph.NodeKindTarget}
	}

	for _, target := range targets {
		fromID, ok := targetIDToNodeID[target.ID]
		if !ok {
			continue
		}

		for _, depTargetID := range target.TargetDependsOn {
			toID, exists := targetIDToNodeID[depTargetID]
			if !exists {
				continue
			}
			edge := graph.Edge{FromID: fromID, ToID: toID, Kind: graph.EdgeKindTarget}
			key := graph.EdgeKey(edge)
			if _, ok := edgeDedup[key]; ok {
				continue
			}
			edgeDedup[key] = struct{}{}
			edges = append(edges, edge)
		}

		for _, product := range target.Products {
			if product.Name == "" {
				continue
			}
			productID := productNodeID(product.PackageIdentity, product.Name)
			if _, ok := nodes[productID]; !ok {
				nodes[productID] = graph.Node{ID: productID, Label: product.Name, Kind: graph.NodeKindExternalProduct}
			}
			edge := graph.Edge{FromID: fromID, ToID: productID, Kind: graph.EdgeKindProduct}
			key := graph.EdgeKey(edge)
			if _, ok := edgeDedup[key]; ok {
				continue
			}
			edgeDedup[key] = struct{}{}
			edges = append(edges, edge)
		}
	}

	g := graph.Graph{Nodes: nodes, Edges: edges}
	g.Edges = graph.SortedEdges(g)
	return g, nil
}
