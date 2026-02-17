package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
	"swift-deps-diagram/internal/manifest"
)

func withManifestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Package.swift")
	if err := os.WriteFile(path, []byte("// fixture"), 0o644); err != nil {
		t.Fatalf("failed to write fixture manifest: %v", err)
	}
	return dir
}

func stubAppDeps(t *testing.T) *string {
	t.Helper()
	oldDump := dumpPackage
	oldDecode := decodeManifest
	oldBuild := buildGraph
	oldMermaid := renderMermaid
	oldDot := renderDot
	oldWrite := writeOutput
	t.Cleanup(func() {
		dumpPackage = oldDump
		decodeManifest = oldDecode
		buildGraph = oldBuild
		renderMermaid = oldMermaid
		renderDot = oldDot
		writeOutput = oldWrite
	})

	dumpPackage = func(context.Context, string) ([]byte, error) { return []byte(`{"name":"X"}`), nil }
	decodeManifest = func([]byte) (manifest.Package, error) { return manifest.Package{}, nil }
	buildGraph = func(manifest.Package, bool) (graph.Graph, error) {
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}
	renderMermaid = func(graph.Graph) (string, error) { return "MERMAID", nil }
	renderDot = func(graph.Graph) (string, error) { return "DOT", nil }

	got := ""
	writeOutput = func(content, _ string, _ io.Writer) error {
		got = content
		return nil
	}
	return &got
}

func TestRunMermaidMode(t *testing.T) {
	dir := withManifestDir(t)
	got := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Format: "mermaid"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if *got != "MERMAID" {
		t.Fatalf("expected MERMAID output, got %q", *got)
	}
}

func TestRunDotMode(t *testing.T) {
	dir := withManifestDir(t)
	got := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Format: "dot"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if *got != "DOT" {
		t.Fatalf("expected DOT output, got %q", *got)
	}
}

func TestRunBothModeSeparator(t *testing.T) {
	dir := withManifestDir(t)
	got := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Format: "both"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if *got != "MERMAID\n\n---\n\nDOT" {
		t.Fatalf("unexpected both output %q", *got)
	}
}

func TestRunInvalidArgs(t *testing.T) {
	err := Run(context.Background(), Options{PackagePath: ".", Format: "bad"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected invalid args error")
	}
	if !apperrors.IsKind(err, apperrors.KindInvalidArgs) {
		t.Fatalf("expected invalid args kind, got %v", err)
	}
}

func TestRunDumpFailure(t *testing.T) {
	dir := withManifestDir(t)
	stubAppDeps(t)
	dumpPackage = func(context.Context, string) ([]byte, error) {
		return nil, apperrors.New(apperrors.KindDumpPackage, "dump failed", errors.New("boom"))
	}

	err := Run(context.Background(), Options{PackagePath: dir, Format: "dot"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected dump error")
	}
	if !apperrors.IsKind(err, apperrors.KindDumpPackage) {
		t.Fatalf("expected dump package kind, got %v", err)
	}
}
