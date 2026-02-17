package graphviz

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
)

func TestWritePNGNoopWhenPathEmpty(t *testing.T) {
	if err := WritePNG(context.Background(), "digraph {}", ""); err != nil {
		t.Fatalf("expected nil error for empty output path, got %v", err)
	}
}

func TestWritePNGGraphvizNotFound(t *testing.T) {
	oldLookPath := lookPath
	oldRunDot := runDot
	lookPath = func(string) (string, error) { return "", errors.New("missing") }
	runDot = func(context.Context, string, string) ([]byte, error) { return nil, nil }
	t.Cleanup(func() {
		lookPath = oldLookPath
		runDot = oldRunDot
	})

	err := WritePNG(context.Background(), "digraph {}", filepath.Join(t.TempDir(), "out.png"))
	if err == nil {
		t.Fatal("expected graphviz missing error")
	}
	if !apperrors.IsKind(err, apperrors.KindGraphvizNotFound) {
		t.Fatalf("expected graphviz not found kind, got %v", err)
	}
}

func TestWritePNGCommandFailureIncludesStderr(t *testing.T) {
	oldLookPath := lookPath
	oldRunDot := runDot
	lookPath = func(string) (string, error) { return "/usr/bin/dot", nil }
	runDot = func(context.Context, string, string) ([]byte, error) {
		return []byte("syntax error near"), errors.New("exit 1")
	}
	t.Cleanup(func() {
		lookPath = oldLookPath
		runDot = oldRunDot
	})

	err := WritePNG(context.Background(), "bad", filepath.Join(t.TempDir(), "out.png"))
	if err == nil {
		t.Fatal("expected render error")
	}
	if !apperrors.IsKind(err, apperrors.KindGraphvizRender) {
		t.Fatalf("expected graphviz render kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "syntax error near") {
		t.Fatalf("expected stderr details in error, got %q", err.Error())
	}
}

func TestWritePNGSuccessCreatesOutputDirectory(t *testing.T) {
	oldLookPath := lookPath
	oldRunDot := runDot
	lookPath = func(string) (string, error) { return "/usr/bin/dot", nil }
	runDot = func(_ context.Context, _ string, outputPath string) ([]byte, error) {
		if err := os.WriteFile(outputPath, []byte("png"), 0o644); err != nil {
			return nil, err
		}
		return nil, nil
	}
	t.Cleanup(func() {
		lookPath = oldLookPath
		runDot = oldRunDot
	})

	outputPath := filepath.Join(t.TempDir(), "nested", "graph.png")
	if err := WritePNG(context.Background(), "digraph {}", outputPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected png file to exist: %v", err)
	}
}
