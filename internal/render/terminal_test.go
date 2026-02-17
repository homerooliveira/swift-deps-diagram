package render

import (
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
)

func TestTerminalRendersASCIIForest(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"target::App":         {ID: "target::App", Label: "App", Kind: graph.NodeKindTarget},
			"target::Core":        {ID: "target::Core", Label: "Core", Kind: graph.NodeKindTarget},
			"target::UI":          {ID: "target::UI", Label: "UI", Kind: graph.NodeKindTarget},
			"target::Tool":        {ID: "target::Tool", Label: "Tool", Kind: graph.NodeKindTarget},
			"pkg::x::Networking":  {ID: "pkg::x::Networking", Label: "Networking", Kind: graph.NodeKindExternalProduct},
			"pkg::x::Alamofire":   {ID: "pkg::x::Alamofire", Label: "Alamofire", Kind: graph.NodeKindExternalProduct},
			"pkg::x::LineBreaker": {ID: "pkg::x::LineBreaker", Label: "Line\nBreaker", Kind: graph.NodeKindExternalProduct},
		},
		Edges: []graph.Edge{
			{FromID: "target::App", ToID: "target::Core", Kind: graph.EdgeKindTarget},
			{FromID: "target::App", ToID: "target::UI", Kind: graph.EdgeKindTarget},
			{FromID: "target::Core", ToID: "pkg::x::Networking", Kind: graph.EdgeKindProduct},
			{FromID: "target::UI", ToID: "target::Core", Kind: graph.EdgeKindTarget},
			{FromID: "target::App", ToID: "pkg::x::Alamofire", Kind: graph.EdgeKindProduct},
			{FromID: "target::App", ToID: "pkg::x::LineBreaker", Kind: graph.EdgeKindProduct},
		},
	}

	out, err := Terminal(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "App\n" +
		"|-- Alamofire\n" +
		"|-- Core\n" +
		"|   \\-- Networking\n" +
		"|-- Line Breaker\n" +
		"\\-- UI\n" +
		"    \\-- Core\n" +
		"        \\-- Networking\n\n" +
		"Tool"
	if out != expected {
		t.Fatalf("unexpected terminal output:\n%s", out)
	}
}

func TestTerminalFallsBackToDependencyRootsWhenNoZeroInDegreeTargetRoots(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"target::A": {ID: "target::A", Label: "A", Kind: graph.NodeKindTarget},
			"target::B": {ID: "target::B", Label: "B", Kind: graph.NodeKindTarget},
		},
		Edges: []graph.Edge{
			{FromID: "target::A", ToID: "target::B", Kind: graph.EdgeKindTarget},
			{FromID: "target::B", ToID: "target::A", Kind: graph.EdgeKindTarget},
		},
	}

	out, err := Terminal(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "A\n" +
		"\\-- B\n" +
		"    \\-- A (*)"
	if out != expected {
		t.Fatalf("unexpected cycle output:\n%s", out)
	}
}

func TestTerminalDeterministicOutput(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"target::App":         {ID: "target::App", Label: "App", Kind: graph.NodeKindTarget},
			"target::Core":        {ID: "target::Core", Label: "Core", Kind: graph.NodeKindTarget},
			"pkg::x::ExternalLib": {ID: "pkg::x::ExternalLib", Label: "ExternalLib", Kind: graph.NodeKindExternalProduct},
		},
		Edges: []graph.Edge{
			{FromID: "target::App", ToID: "target::Core", Kind: graph.EdgeKindTarget},
			{FromID: "target::App", ToID: "pkg::x::ExternalLib", Kind: graph.EdgeKindProduct},
		},
	}

	first, err := Terminal(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := Terminal(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != second {
		t.Fatalf("expected deterministic output")
	}
}

func TestTerminalUsesByNameTargetLinksForRootDetection(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			"target::swift-deps-diagram": {ID: "target::swift-deps-diagram", Label: "swift-deps-diagram", Kind: graph.NodeKindTarget},
			"target::CLI":                {ID: "target::CLI", Label: "CLI", Kind: graph.NodeKindTarget},
			"target::Runtime":            {ID: "target::Runtime", Label: "Runtime", Kind: graph.NodeKindTarget},
		},
		Edges: []graph.Edge{
			{FromID: "target::swift-deps-diagram", ToID: "target::CLI", Kind: graph.EdgeKindByName},
			{FromID: "target::CLI", ToID: "target::Runtime", Kind: graph.EdgeKindByName},
		},
	}

	out, err := Terminal(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "swift-deps-diagram\n\\-- CLI\n    \\-- Runtime"
	if out != expected {
		t.Fatalf("unexpected terminal output:\n%s", out)
	}
}

func TestTerminalEmptyGraph(t *testing.T) {
	out, err := Terminal(graph.Graph{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "(empty)" {
		t.Fatalf("expected (empty), got %q", out)
	}
}

func TestTerminalIncludesIsolatedTargetsWithoutDependencies(t *testing.T) {
	out, err := Terminal(graph.Graph{
		Nodes: map[string]graph.Node{
			"target::Solo": {ID: "target::Solo", Label: "Solo", Kind: graph.NodeKindTarget},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Solo" {
		t.Fatalf("expected Solo, got %q", out)
	}
}

func TestTerminalErrorsOnMissingNodes(t *testing.T) {
	_, err := Terminal(graph.Graph{
		Nodes: map[string]graph.Node{
			"target::App": {ID: "target::App", Label: "App", Kind: graph.NodeKindTarget},
		},
		Edges: []graph.Edge{
			{FromID: "target::App", ToID: "target::Missing", Kind: graph.EdgeKindTarget},
		},
	})
	if err == nil {
		t.Fatal("expected error for missing node")
	}
	if !apperrors.IsKind(err, apperrors.KindRuntime) {
		t.Fatalf("expected runtime kind, got %v", err)
	}
}
