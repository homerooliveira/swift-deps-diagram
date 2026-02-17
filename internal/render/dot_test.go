package render

import (
	"strings"
	"testing"

	"swift-deps-diagram/internal/graph"
)

func TestDotIncludesGraphHeaderAndFooter(t *testing.T) {
	out, err := Dot(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "digraph dependencies {") {
		t.Fatalf("expected digraph header, got %q", out)
	}
	if !strings.HasSuffix(out, "}") {
		t.Fatalf("expected closing brace, got %q", out)
	}
}

func TestDotRendersAllNodesAndEdges(t *testing.T) {
	out, err := Dot(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, part := range []string{"\"target::App\"", "\"target::Core\"", "->"} {
		if !strings.Contains(out, part) {
			t.Fatalf("missing output segment %q", part)
		}
	}
}

func TestDotQuotesIDsAndLabels(t *testing.T) {
	g := graph.Graph{
		Nodes: map[string]graph.Node{
			`id"1`: {ID: `id"1`, Label: `A"pp`, Kind: graph.NodeKindTarget},
		},
	}
	out, err := Dot(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"id\"1"`) {
		t.Fatalf("expected quoted escaped ID, got %q", out)
	}
	if !strings.Contains(out, `label="A\"pp"`) {
		t.Fatalf("expected quoted escaped label, got %q", out)
	}
}

func TestDotAppliesNodeStylesByKind(t *testing.T) {
	out, err := Dot(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "shape=box") {
		t.Fatalf("expected target style, got %q", out)
	}
	if !strings.Contains(out, "shape=ellipse,style=dashed") {
		t.Fatalf("expected external style, got %q", out)
	}
}

func TestDotDeterministicOutput(t *testing.T) {
	a, err := Dot(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := Dot(sampleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a != b {
		t.Fatalf("expected deterministic output")
	}
}
