package xcodegraph

import (
	"testing"

	"swift-deps-diagram/internal/graph"
	"swift-deps-diagram/internal/xcodeproj"
)

func TestBuildCreatesTargetAndProductEdges(t *testing.T) {
	project := xcodeproj.Project{Targets: []xcodeproj.Target{
		{
			ID:              "T_APP",
			Name:            "App",
			TargetDependsOn: []string{"T_CORE"},
			Products: []xcodeproj.PackageProduct{{
				Name:            "Alamofire",
				PackageIdentity: "alamofire",
			}},
		},
		{
			ID:   "T_CORE",
			Name: "Core",
		},
	}}

	g, err := Build(project, false)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	if _, ok := g.Nodes["target::App"]; !ok {
		t.Fatal("missing App target node")
	}
	if _, ok := g.Nodes["target::Core"]; !ok {
		t.Fatal("missing Core target node")
	}
	if _, ok := g.Nodes["pkg::alamofire::Alamofire"]; !ok {
		t.Fatal("missing product node")
	}

	targetEdgeFound := false
	productEdgeFound := false
	for _, edge := range g.Edges {
		if edge == (graph.Edge{FromID: "target::App", ToID: "target::Core", Kind: graph.EdgeKindTarget}) {
			targetEdgeFound = true
		}
		if edge == (graph.Edge{FromID: "target::App", ToID: "pkg::alamofire::Alamofire", Kind: graph.EdgeKindProduct}) {
			productEdgeFound = true
		}
	}
	if !targetEdgeFound {
		t.Fatal("missing target edge")
	}
	if !productEdgeFound {
		t.Fatal("missing product edge")
	}
}

func TestBuildRespectsIncludeTests(t *testing.T) {
	project := xcodeproj.Project{Targets: []xcodeproj.Target{
		{ID: "T_APP", Name: "App", ProductType: "com.apple.product-type.application"},
		{ID: "T_TESTS", Name: "AppTests", ProductType: "com.apple.product-type.bundle.unit-test"},
	}}

	withoutTests, err := Build(project, false)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if _, ok := withoutTests.Nodes["target::AppTests"]; ok {
		t.Fatal("did not expect test target when includeTests=false")
	}

	withTests, err := Build(project, true)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if _, ok := withTests.Nodes["target::AppTests"]; !ok {
		t.Fatal("expected test target when includeTests=true")
	}
}
