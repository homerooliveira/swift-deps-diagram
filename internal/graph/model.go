package graph

import "sort"

type NodeKind string

const (
	NodeKindTarget          NodeKind = "target"
	NodeKindExternalProduct NodeKind = "external_product"
)

type EdgeKind string

const (
	EdgeKindTarget  EdgeKind = "target"
	EdgeKindProduct EdgeKind = "product"
	EdgeKindByName  EdgeKind = "by_name"
)

type Node struct {
	ID    string
	Label string
	Kind  NodeKind
}

type Edge struct {
	FromID string
	ToID   string
	Kind   EdgeKind
}

type Graph struct {
	Nodes map[string]Node
	Edges []Edge
}

func EdgeKey(e Edge) string {
	return string(e.Kind) + "|" + e.FromID + "|" + e.ToID
}

func SortedNodeIDs(g Graph) []string {
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func SortedEdges(g Graph) []Edge {
	edges := make([]Edge, len(g.Edges))
	copy(edges, g.Edges)
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].FromID != edges[j].FromID {
			return edges[i].FromID < edges[j].FromID
		}
		if edges[i].ToID != edges[j].ToID {
			return edges[i].ToID < edges[j].ToID
		}
		return edges[i].Kind < edges[j].Kind
	})
	return edges
}
