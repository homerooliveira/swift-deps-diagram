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
	if opts.Mode != "auto" {
		t.Fatalf("expected default mode auto, got %q", opts.Mode)
	}
	if opts.Format != "png" {
		t.Fatalf("expected default format png, got %q", opts.Format)
	}
	if opts.Output != "" {
		t.Fatalf("expected default output empty, got %q", opts.Output)
	}
	if opts.Verbose {
		t.Fatalf("expected default verbose false")
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

func TestParseFlagsInvalidMode(t *testing.T) {
	var stderr bytes.Buffer
	_, err := parseFlags([]string{"--mode", "bad"}, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !apperrors.IsKind(err, apperrors.KindInvalidArgs) {
		t.Fatalf("expected invalid args kind, got %v", err)
	}
}

func TestParseFlagsVerbose(t *testing.T) {
	var stderr bytes.Buffer
	opts, err := parseFlags([]string{"--verbose"}, &stderr)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !opts.Verbose {
		t.Fatal("expected verbose=true")
	}
}

func TestParseFlagsRejectsPNGOutput(t *testing.T) {
	var stderr bytes.Buffer
	_, err := parseFlags([]string{"--png-output", "diagram.png"}, &stderr)
	if err == nil {
		t.Fatal("expected parse error for removed --png-output flag")
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

func TestExecutePassesVerboseToApp(t *testing.T) {
	oldRun := runApp
	defer func() { runApp = oldRun }()

	var got app.Options
	runApp = func(_ context.Context, opts app.Options, _ io.Writer) error {
		got = opts
		return nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := execute([]string{"--verbose", "--format", "dot", "--output", "deps.dot"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !got.Verbose {
		t.Fatal("expected verbose=true in app options")
	}
	if got.Format != "dot" {
		t.Fatalf("expected format dot, got %q", got.Format)
	}
}

func TestHelpTextIncludesFlags(t *testing.T) {
	var stderr bytes.Buffer
	_, err := parseFlags([]string{"-h"}, &stderr)
	if err == nil {
		t.Fatal("expected help path to return error")
	}
	output := stderr.String()
	for _, needle := range []string{"-path", "-project", "-workspace", "-mode", "-format", "-output", "-verbose", "-include-tests"} {
		if !bytes.Contains([]byte(output), []byte(needle)) {
			t.Fatalf("help output missing %s", needle)
		}
	}
}
