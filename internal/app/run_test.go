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
	"swift-deps-diagram/internal/inputresolve"
	"swift-deps-diagram/internal/manifest"
	"swift-deps-diagram/internal/xcodeproj"
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
	oldResolve := resolveInput
	oldDump := dumpPackage
	oldDecode := decodeManifest
	oldBuild := buildGraph
	oldLoadXcode := loadXcodeProject
	oldBuildXcode := buildXcodeGraph
	oldMermaid := renderMermaid
	oldDot := renderDot
	oldWrite := writeOutput
	oldWritePNG := writePNG
	t.Cleanup(func() {
		resolveInput = oldResolve
		dumpPackage = oldDump
		decodeManifest = oldDecode
		buildGraph = oldBuild
		loadXcodeProject = oldLoadXcode
		buildXcodeGraph = oldBuildXcode
		renderMermaid = oldMermaid
		renderDot = oldDot
		writeOutput = oldWrite
		writePNG = oldWritePNG
	})

	resolveInput = func(req inputresolve.Request) (inputresolve.Resolved, error) {
		return inputresolve.Resolved{Mode: inputresolve.ModeSPM, PackagePath: req.Path}, nil
	}
	dumpPackage = func(context.Context, string) ([]byte, error) { return []byte(`{"name":"X"}`), nil }
	decodeManifest = func([]byte) (manifest.Package, error) { return manifest.Package{}, nil }
	buildGraph = func(manifest.Package, bool) (graph.Graph, error) {
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}
	loadXcodeProject = func(context.Context, string) (xcodeproj.Project, error) { return xcodeproj.Project{}, nil }
	buildXcodeGraph = func(xcodeproj.Project, bool) (graph.Graph, error) {
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}
	renderMermaid = func(graph.Graph) (string, error) { return "MERMAID", nil }
	renderDot = func(graph.Graph) (string, error) { return "DOT", nil }

	got := ""
	writeOutput = func(content, _ string, _ io.Writer) error {
		got = content
		return nil
	}
	writePNG = func(context.Context, string, string) error { return nil }
	return &got
}

func TestRunMermaidMode(t *testing.T) {
	dir := withManifestDir(t)
	got := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "mermaid"}, &bytes.Buffer{})
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

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot"}, &bytes.Buffer{})
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

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "both"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if *got != "MERMAID\n\n---\n\nDOT" {
		t.Fatalf("unexpected both output %q", *got)
	}
}

func TestRunInvalidArgs(t *testing.T) {
	err := Run(context.Background(), Options{PackagePath: ".", Mode: "auto", Format: "bad"}, &bytes.Buffer{})
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

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected dump error")
	}
	if !apperrors.IsKind(err, apperrors.KindDumpPackage) {
		t.Fatalf("expected dump package kind, got %v", err)
	}
}

func TestRunPNGOutputRequested(t *testing.T) {
	dir := withManifestDir(t)
	stubAppDeps(t)

	pngCalled := false
	var gotPath string
	var gotDot string
	writePNG = func(_ context.Context, dotSource, outputPath string) error {
		pngCalled = true
		gotPath = outputPath
		gotDot = dotSource
		return nil
	}

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "mermaid", PNGOutput: "out/diagram.png"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !pngCalled {
		t.Fatal("expected png generation to be called")
	}
	if gotPath != "out/diagram.png" {
		t.Fatalf("unexpected png output path %q", gotPath)
	}
	if gotDot != "DOT" {
		t.Fatalf("expected DOT source passed to png generator, got %q", gotDot)
	}
}

func TestRunXcodeModeUsesXcodePipeline(t *testing.T) {
	dir := withManifestDir(t)
	got := stubAppDeps(t)

	resolveInput = func(inputresolve.Request) (inputresolve.Resolved, error) {
		return inputresolve.Resolved{Mode: inputresolve.ModeXcode, ProjectPath: "/tmp/App.xcodeproj"}, nil
	}
	loadCalled := false
	buildCalled := false
	loadXcodeProject = func(context.Context, string) (xcodeproj.Project, error) {
		loadCalled = true
		return xcodeproj.Project{Targets: []xcodeproj.Target{{ID: "A", Name: "App"}}}, nil
	}
	buildXcodeGraph = func(xcodeproj.Project, bool) (graph.Graph, error) {
		buildCalled = true
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !loadCalled {
		t.Fatal("expected xcode loader to be called")
	}
	if !buildCalled {
		t.Fatal("expected xcode graph builder to be called")
	}
	if *got != "DOT" {
		t.Fatalf("expected DOT output, got %q", *got)
	}
}
