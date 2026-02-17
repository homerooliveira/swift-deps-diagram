package graph

import (
	"testing"

	"swift-deps-diagram/internal/manifest"
	"swift-deps-diagram/internal/testutil"
)

func mustDecodePackage(t *testing.T, fixture string) manifest.Package {
	t.Helper()
	pkg, err := manifest.Decode(testutil.ReadFixture(t, fixture))
	if err != nil {
		t.Fatalf("decode fixture failed: %v", err)
	}
	return pkg
}

func TestBuildGraphLocalTargetDependencies(t *testing.T) {
	pkg := mustDecodePackage(t, "simple-local.json")
	g, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if _, ok := g.Nodes[targetNodeID("App")]; !ok {
		t.Fatal("missing App target node")
	}
	if _, ok := g.Nodes[targetNodeID("Core")]; !ok {
		t.Fatal("missing Core target node")
	}
	found := false
	for _, e := range g.Edges {
		if e.FromID == targetNodeID("App") && e.ToID == targetNodeID("Core") && e.Kind == EdgeKindTarget {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected App -> Core target edge")
	}
}

func TestBuildGraphProductDependencies(t *testing.T) {
	pkg := mustDecodePackage(t, "product-and-byname.json")
	g, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	id := productNodeID("Alamofire", "alamofire")
	n, ok := g.Nodes[id]
	if !ok {
		t.Fatalf("expected product node %q", id)
	}
	if n.Kind != NodeKindExternalProduct {
		t.Fatalf("expected external product kind, got %s", n.Kind)
	}
}

func TestBuildGraphByNameResolution(t *testing.T) {
	pkg := mustDecodePackage(t, "product-and-byname.json")
	g, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	hasLocalByName := false
	hasExternalByName := false
	for _, e := range g.Edges {
		if e.Kind != EdgeKindByName {
			continue
		}
		if e.ToID == targetNodeID("Core") {
			hasLocalByName = true
		}
		if e.ToID == byNameNodeID("SomeExternal") {
			hasExternalByName = true
		}
	}
	if !hasLocalByName {
		t.Fatal("expected byName local resolution to Core")
	}
	if !hasExternalByName {
		t.Fatal("expected byName external fallback")
	}
}

func TestBuildGraphExcludeAndIncludeTests(t *testing.T) {
	pkg := mustDecodePackage(t, "product-and-byname.json")

	withoutTests, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("build without tests failed: %v", err)
	}
	if _, ok := withoutTests.Nodes[targetNodeID("AppTests")]; ok {
		t.Fatal("did not expect AppTests node when includeTests=false")
	}

	withTests, err := Build(pkg, true)
	if err != nil {
		t.Fatalf("build with tests failed: %v", err)
	}
	if _, ok := withTests.Nodes[targetNodeID("AppTests")]; !ok {
		t.Fatal("expected AppTests node when includeTests=true")
	}
}

func TestBuildGraphDeduplicatesEdges(t *testing.T) {
	pkg := manifest.Package{
		Name: "Sample",
		Targets: []manifest.Target{
			{
				Name: "App",
				Type: "regular",
				Dependencies: []manifest.TargetDependency{
					{Kind: manifest.DependencyKindTarget, Name: "Core"},
					{Kind: manifest.DependencyKindTarget, Name: "Core"},
				},
			},
			{Name: "Core", Type: "regular"},
		},
	}
	g, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	count := 0
	for _, e := range g.Edges {
		if e.FromID == targetNodeID("App") && e.ToID == targetNodeID("Core") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 deduped edge, got %d", count)
	}
}

func TestBuildGraphDeterministicOrdering(t *testing.T) {
	pkg := mustDecodePackage(t, "product-and-byname.json")
	a, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("first build failed: %v", err)
	}
	b, err := Build(pkg, false)
	if err != nil {
		t.Fatalf("second build failed: %v", err)
	}

	aEdges := SortedEdges(a)
	bEdges := SortedEdges(b)
	if len(aEdges) != len(bEdges) {
		t.Fatalf("edge count mismatch: %d vs %d", len(aEdges), len(bEdges))
	}
	for i := range aEdges {
		if aEdges[i] != bEdges[i] {
			t.Fatalf("edge order mismatch at %d: %#v vs %#v", i, aEdges[i], bEdges[i])
		}
	}
}
