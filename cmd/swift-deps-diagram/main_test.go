package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"swift-deps-diagram/internal/app"
	apperrors "swift-deps-diagram/internal/errors"
)

func TestParseFlagsDefaults(t *testing.T) {
	var stderr bytes.Buffer
	opts, err := parseFlags([]string{}, &stderr)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if opts.Path != "." {
		t.Fatalf("expected default path '.', got %q", opts.Path)
	}
	if opts.Format != "both" {
		t.Fatalf("expected default format both, got %q", opts.Format)
	}
	if opts.Output != "" {
		t.Fatalf("expected default output empty, got %q", opts.Output)
	}
	if opts.IncludeTests {
		t.Fatalf("expected default include-tests false")
	}
}

func TestParseFlagsInvalidFormat(t *testing.T) {
	var stderr bytes.Buffer
	_, err := parseFlags([]string{"--format", "bad"}, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !apperrors.IsKind(err, apperrors.KindInvalidArgs) {
		t.Fatalf("expected invalid args kind, got %v", err)
	}
}

func TestExecuteMapsRuntimeErrorToExitCode2(t *testing.T) {
	oldRun := runApp
	runApp = func(_ context.Context, _ app.Options, _ io.Writer) error {
		return apperrors.New(apperrors.KindRuntime, "runtime", errors.New("boom"))
	}
	defer func() { runApp = oldRun }()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := execute([]string{}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

func TestHelpTextIncludesFlags(t *testing.T) {
	var stderr bytes.Buffer
	_, err := parseFlags([]string{"-h"}, &stderr)
	if err == nil {
		t.Fatal("expected help path to return error")
	}
	output := stderr.String()
	for _, needle := range []string{"-path", "-format", "-output", "-include-tests"} {
		if !bytes.Contains([]byte(output), []byte(needle)) {
			t.Fatalf("help output missing %s", needle)
		}
	}
}
