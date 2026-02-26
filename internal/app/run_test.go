package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"swift-deps-diagram/internal/bazel"
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

type appHarness struct {
	textOutput  string
	pngPath     string
	pngDot      string
	logMessages []string
}

func stubAppDeps(t *testing.T) *appHarness {
	t.Helper()
	h := &appHarness{}

	oldResolve := resolveInput
	oldDump := dumpPackage
	oldDecode := decodeManifest
	oldBuild := buildGraph
	oldLoadXcode := loadXcodeProject
	oldGenerateTuist := generateTuistProject
	oldBuildXcode := buildXcodeGraph
	oldLoadBazel := loadBazelWorkspace
	oldBuildBazel := buildBazelGraph
	oldMermaid := renderMermaid
	oldDot := renderDot
	oldTerminal := renderTerminal
	oldWrite := writeOutput
	oldWritePNG := writePNG
	oldLogInfof := logInfof
	t.Cleanup(func() {
		resolveInput = oldResolve
		dumpPackage = oldDump
		decodeManifest = oldDecode
		buildGraph = oldBuild
		loadXcodeProject = oldLoadXcode
		generateTuistProject = oldGenerateTuist
		buildXcodeGraph = oldBuildXcode
		loadBazelWorkspace = oldLoadBazel
		buildBazelGraph = oldBuildBazel
		renderMermaid = oldMermaid
		renderDot = oldDot
		renderTerminal = oldTerminal
		writeOutput = oldWrite
		writePNG = oldWritePNG
		logInfof = oldLogInfof
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
	generateTuistProject = func(context.Context, string) error { return nil }
	buildXcodeGraph = func(xcodeproj.Project, bool) (graph.Graph, error) {
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}
	loadBazelWorkspace = func(context.Context, string, string) (bazel.Workspace, error) {
		return bazel.Workspace{}, nil
	}
	buildBazelGraph = func(bazel.Workspace, bool) (graph.Graph, error) {
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}
	renderMermaid = func(graph.Graph) (string, error) { return "MERMAID", nil }
	renderDot = func(graph.Graph) (string, error) { return "DOT", nil }
	renderTerminal = func(graph.Graph) (string, error) { return "TERMINAL", nil }
	writeOutput = func(content, _ string, _ io.Writer) error {
		h.textOutput = content
		return nil
	}
	writePNG = func(_ context.Context, dotSource, outputPath string) error {
		h.pngPath = outputPath
		h.pngDot = dotSource
		return nil
	}
	logInfof = func(format string, args ...interface{}) {
		h.logMessages = append(h.logMessages, fmt.Sprintf(format, args...))
	}

	return h
}

func TestRunMermaidModeWritesText(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "mermaid"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if h.textOutput != "MERMAID" {
		t.Fatalf("expected MERMAID output, got %q", h.textOutput)
	}
	if h.pngPath != "" {
		t.Fatalf("expected no png generation, got %q", h.pngPath)
	}
}

func TestRunDotModeWritesText(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if h.textOutput != "DOT" {
		t.Fatalf("expected DOT output, got %q", h.textOutput)
	}
}

func TestRunTerminalModeWritesText(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "terminal"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if h.textOutput != "TERMINAL" {
		t.Fatalf("expected TERMINAL output, got %q", h.textOutput)
	}
}

func TestRunPNGModeUsesDefaultOutputPath(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "png"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if h.pngPath != "deps.png" {
		t.Fatalf("expected default png path deps.png, got %q", h.pngPath)
	}
	if h.pngDot != "DOT" {
		t.Fatalf("expected DOT source for png, got %q", h.pngDot)
	}
}

func TestRunPNGModeUsesCustomOutputPath(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "png", OutputPath: "out/custom.png"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if h.pngPath != "out/custom.png" {
		t.Fatalf("expected custom png path, got %q", h.pngPath)
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

func TestRunVerboseMessageRules(t *testing.T) {
	dir := withManifestDir(t)

	t.Run("mermaid_stdout_no_message", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "mermaid", Verbose: true}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.logMessages) != 0 {
			t.Fatalf("expected no verbose messages, got %#v", h.logMessages)
		}
	})

	t.Run("terminal_stdout_no_message", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "terminal", Verbose: true}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.logMessages) != 0 {
			t.Fatalf("expected no verbose messages, got %#v", h.logMessages)
		}
	})

	t.Run("mermaid_file_message", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "mermaid", OutputPath: "deps.mmd", Verbose: true}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.logMessages) != 1 || h.logMessages[0] != "generated mermaid content at deps.mmd" {
			t.Fatalf("unexpected messages %#v", h.logMessages)
		}
	})

	t.Run("dot_file_message", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot", OutputPath: "deps.dot", Verbose: true}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.logMessages) != 1 || h.logMessages[0] != "generated dot content at deps.dot" {
			t.Fatalf("unexpected messages %#v", h.logMessages)
		}
	})

	t.Run("terminal_file_message", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "terminal", OutputPath: "deps.tree", Verbose: true}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.logMessages) != 1 || h.logMessages[0] != "generated terminal content at deps.tree" {
			t.Fatalf("unexpected messages %#v", h.logMessages)
		}
	})

	t.Run("png_default_message", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "png", Verbose: true}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := fmt.Sprintf("generated png using dot format at %s", absolutePath("deps.png"))
		if len(h.logMessages) != 1 || h.logMessages[0] != expected {
			t.Fatalf("unexpected messages %#v", h.logMessages)
		}
	})

	t.Run("png_non_verbose_still_logs_output_path", func(t *testing.T) {
		h := stubAppDeps(t)
		err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "png", OutputPath: "deps.png", Verbose: false}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := fmt.Sprintf("generated png using dot format at %s", absolutePath("deps.png"))
		if len(h.logMessages) != 1 || h.logMessages[0] != expected {
			t.Fatalf("unexpected messages %#v", h.logMessages)
		}
	})
}

func TestRunXcodeModeUsesXcodePipeline(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

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
	if h.textOutput != "DOT" {
		t.Fatalf("expected DOT output, got %q", h.textOutput)
	}
}

func TestRunXcodeModeGeneratesTuistProject(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	resolveCalls := 0
	resolveInput = func(req inputresolve.Request) (inputresolve.Resolved, error) {
		resolveCalls++
		if resolveCalls == 1 {
			return inputresolve.Resolved{Mode: inputresolve.ModeXcode, TuistPath: "/tmp/tuist-app"}, nil
		}
		if req.Path != "/tmp/tuist-app" || req.Mode != inputresolve.ModeXcode {
			t.Fatalf("unexpected second resolve request: %#v", req)
		}
		return inputresolve.Resolved{Mode: inputresolve.ModeXcode, ProjectPath: "/tmp/tuist-app/App.xcodeproj"}, nil
	}

	generated := false
	generateTuistProject = func(_ context.Context, path string) error {
		generated = true
		if path != "/tmp/tuist-app" {
			t.Fatalf("unexpected tuist path: %s", path)
		}
		return nil
	}

	loaded := false
	loadXcodeProject = func(_ context.Context, projectPath string) (xcodeproj.Project, error) {
		loaded = true
		if projectPath != "/tmp/tuist-app/App.xcodeproj" {
			t.Fatalf("unexpected project path: %s", projectPath)
		}
		return xcodeproj.Project{Targets: []xcodeproj.Target{{ID: "A", Name: "App"}}}, nil
	}

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !generated {
		t.Fatal("expected tuist generation to run")
	}
	if !loaded {
		t.Fatal("expected xcode loader to run")
	}
	if h.textOutput != "DOT" {
		t.Fatalf("expected DOT output, got %q", h.textOutput)
	}
}

func TestRunBazelModeUsesBazelPipeline(t *testing.T) {
	dir := withManifestDir(t)
	h := stubAppDeps(t)

	resolveInput = func(inputresolve.Request) (inputresolve.Resolved, error) {
		return inputresolve.Resolved{
			Mode:               inputresolve.ModeBazel,
			BazelWorkspacePath: "/tmp/workspace",
			BazelTargets:       "//app:cli",
		}, nil
	}
	loadCalled := false
	buildCalled := false
	loadBazelWorkspace = func(_ context.Context, path, targets string) (bazel.Workspace, error) {
		loadCalled = true
		if path != "/tmp/workspace" {
			t.Fatalf("unexpected workspace path %q", path)
		}
		if targets != "//app:cli" {
			t.Fatalf("unexpected targets %q", targets)
		}
		return bazel.Workspace{Targets: []bazel.Target{{Label: "//app:cli", Kind: "swift_binary"}}}, nil
	}
	buildBazelGraph = func(workspace bazel.Workspace, includeTests bool) (graph.Graph, error) {
		buildCalled = true
		return graph.Graph{Nodes: map[string]graph.Node{}, Edges: []graph.Edge{}}, nil
	}

	err := Run(context.Background(), Options{PackagePath: dir, Mode: "auto", Format: "dot"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !loadCalled {
		t.Fatal("expected bazel loader to be called")
	}
	if !buildCalled {
		t.Fatal("expected bazel graph builder to be called")
	}
	if h.textOutput != "DOT" {
		t.Fatalf("expected DOT output, got %q", h.textOutput)
	}
}
