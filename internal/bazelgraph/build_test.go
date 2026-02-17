package bazelgraph

import (
	"testing"

	"swift-deps-diagram/internal/bazel"
	"swift-deps-diagram/internal/graph"
)

func TestBuildIncludesMixedRuleKinds(t *testing.T) {
	workspace := bazel.Workspace{
		Targets: []bazel.Target{
			{Label: "//app:bin", Kind: "swift_binary"},
			{Label: "//app:lib", Kind: "swift_library"},
			{Label: "//tools:gen", Kind: "genrule"},
		},
	}

	g, err := Build(workspace, false)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	for _, label := range []string{"//app:bin", "//app:lib", "//tools:gen"} {
		if _, ok := g.Nodes["target::"+label]; !ok {
			t.Fatalf("missing node for %s", label)
		}
	}
}

func TestBuildMapsLocalAndExternalDeps(t *testing.T) {
	workspace := bazel.Workspace{
		Targets: []bazel.Target{
			{
				Label: "//app:bin",
				Kind:  "swift_binary",
				Deps:  []string{"//app:lib", "@repo//pkg:network"},
			},
			{
				Label: "//app:lib",
				Kind:  "swift_library",
			},
		},
	}

	g, err := Build(workspace, false)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	if _, ok := g.Nodes["target:://app:lib"]; !ok {
		t.Fatal("missing local dep node")
	}
	if _, ok := g.Nodes["external::@repo//pkg:network"]; !ok {
		t.Fatal("missing external dep node")
	}

	localFound := false
	externalFound := false
	for _, edge := range g.Edges {
		if edge == (graph.Edge{FromID: "target:://app:bin", ToID: "target:://app:lib", Kind: graph.EdgeKindTarget}) {
			localFound = true
		}
		if edge == (graph.Edge{FromID: "target:://app:bin", ToID: "external::@repo//pkg:network", Kind: graph.EdgeKindProduct}) {
			externalFound = true
		}
	}
	if !localFound {
		t.Fatal("missing local edge")
	}
	if !externalFound {
		t.Fatal("missing external edge")
	}
}

func TestBuildRespectsIncludeTests(t *testing.T) {
	workspace := bazel.Workspace{
		Targets: []bazel.Target{
			{Label: "//app:bin", Kind: "swift_binary", Deps: []string{"//app:lib", "//app:bin_test"}},
			{Label: "//app:lib", Kind: "swift_library"},
			{Label: "//app:bin_test", Kind: "swift_test"},
		},
	}

	withoutTests, err := Build(workspace, false)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if _, ok := withoutTests.Nodes["target:://app:bin_test"]; ok {
		t.Fatal("did not expect test node when includeTests=false")
	}
	for _, edge := range withoutTests.Edges {
		if edge.ToID == "target:://app:bin_test" {
			t.Fatal("did not expect edge to test node when includeTests=false")
		}
	}

	withTests, err := Build(workspace, true)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if _, ok := withTests.Nodes["target:://app:bin_test"]; !ok {
		t.Fatal("expected test node when includeTests=true")
	}
}

func TestBuildDeterministicAndDeduped(t *testing.T) {
	workspace := bazel.Workspace{
		Targets: []bazel.Target{
			{
				Label: "//app:b",
				Kind:  "swift_binary",
				Deps:  []string{"//app:a", "//app:a", "@repo//pkg:z", "@repo//pkg:z"},
			},
			{Label: "//app:a", Kind: "swift_library"},
		},
	}

	g, err := Build(workspace, false)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	if len(g.Edges) != 2 {
		t.Fatalf("expected 2 deduped edges, got %d: %#v", len(g.Edges), g.Edges)
	}
	if g.Edges[0] != (graph.Edge{FromID: "target:://app:b", ToID: "external::@repo//pkg:z", Kind: graph.EdgeKindProduct}) {
		t.Fatalf("unexpected first edge order: %#v", g.Edges)
	}
	if g.Edges[1] != (graph.Edge{FromID: "target:://app:b", ToID: "target:://app:a", Kind: graph.EdgeKindTarget}) {
		t.Fatalf("unexpected second edge order: %#v", g.Edges)
	}
}
