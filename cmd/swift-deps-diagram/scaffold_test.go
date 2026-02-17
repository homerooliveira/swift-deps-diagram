package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestProjectStructure(t *testing.T) {
	root := repoRoot(t)
	requiredDirs := []string{
		"cmd/swift-deps-diagram",
		"internal/app",
		"internal/swiftpm",
		"internal/manifest",
		"internal/graph",
		"internal/render",
		"internal/output",
		"internal/errors",
		"internal/bazel",
		"internal/bazelgraph",
		"testdata/fixtures",
	}

	for _, dir := range requiredDirs {
		full := filepath.Join(root, dir)
		info, err := os.Stat(full)
		if err != nil {
			t.Fatalf("missing required dir %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("path is not a directory: %s", dir)
		}
	}
}
