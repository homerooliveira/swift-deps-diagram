package render

import (
	"strings"
	"testing"

	"swift-deps-diagram/internal/graph"
)

func sampleGraph() graph.Graph {
	return graph.Graph{
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
}

func TestMermaidIncludesFlowchartHeader(t *testing.T) {
	out, err := Mermaid(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "flowchart TD") {
		t.Fatalf("expected flowchart header, got %q", out)
	}
}

func TestMermaidRendersAllNodesAndEdges(t *testing.T) {
	out, err := Mermaid(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, part := range []string{"[\"App\"]", "[\"Core\"]", "[\"ExternalLib\"]", "-->"} {
		if !strings.Contains(out, part) {
			t.Fatalf("missing output segment %q", part)
		}
	}
}

func TestMermaidEscapesLabels(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{"target::App": {ID: "target::App", Label: "A\"pp`\nName", Kind: graph.NodeKindTarget}},
	}
	out, err := Mermaid(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "`") {
		t.Fatalf("expected backticks removed, got %q", out)
	}
	if !strings.Contains(out, "A\\\"pp Name") {
		t.Fatalf("expected escaped label, got %q", out)
	}
}

func TestMermaidUsesDeterministicNodeIDs(t *testing.T) {
	out, err := Mermaid(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "n1") || !strings.Contains(out, "n2") {
		t.Fatalf("expected deterministic node IDs, got %q", out)
	}
}

func TestMermaidDeterministicOutput(t *testing.T) {
	a, err := Mermaid(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := Mermaid(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a != b {
		t.Fatalf("expected deterministic output")
	}
}
