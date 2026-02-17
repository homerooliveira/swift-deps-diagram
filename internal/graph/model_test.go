package graph

import "testing"

func TestNodeKindValuesStable(t *testing.T) {
	if NodeKindTarget != "target" {
		t.Fatalf("unexpected NodeKindTarget value: %q", NodeKindTarget)
	}
	if NodeKindExternalProduct != "external_product" {
		t.Fatalf("unexpected NodeKindExternalProduct value: %q", NodeKindExternalProduct)
	}
}

func TestEdgeKindValuesStable(t *testing.T) {
	if EdgeKindTarget != "target" {
		t.Fatalf("unexpected EdgeKindTarget value: %q", EdgeKindTarget)
	}
	if EdgeKindProduct != "product" {
		t.Fatalf("unexpected EdgeKindProduct value: %q", EdgeKindProduct)
	}
	if EdgeKindByName != "by_name" {
		t.Fatalf("unexpected EdgeKindByName value: %q", EdgeKindByName)
	}
}

func TestGraphEdgeDedupKey(t *testing.T) {
	e := Edge{FromID: "a", ToID: "b", Kind: EdgeKindTarget}
	if EdgeKey(e) != EdgeKey(e) {
		t.Fatal("expected same key for same edge")
	}
}

func TestGraphSortOrder(t *testing.T) {
	g := Graph{
		Nodes: map[string]Node{
			"b": {ID: "b"},
			"a": {ID: "a"},
		},
		Edges: []Edge{
			{FromID: "b", ToID: "a", Kind: EdgeKindTarget},
			{FromID: "a", ToID: "b", Kind: EdgeKindTarget},
		},
	}
	nodes := SortedNodeIDs(g)
	if len(nodes) != 2 || nodes[0] != "a" || nodes[1] != "b" {
		t.Fatalf("unexpected node order: %#v", nodes)
	}
	edges := SortedEdges(g)
	if len(edges) != 2 || edges[0].FromID != "a" {
		t.Fatalf("unexpected edge order: %#v", edges)
	}
}
