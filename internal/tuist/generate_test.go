package tuist

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

func TestGenerateSuccessUsesNoOpenFlag(t *testing.T) {
	oldLookPath := lookPath
	oldRun := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRun
	})

	lookPath = func(file string) (string, error) {
		if file != "tuist" {
			t.Fatalf("expected lookup for tuist, got %s", file)
		}
		return "/usr/bin/tuist", nil
	}

	called := false
	runCommand = func(_ context.Context, _ string, name string, args ...string) ([]byte, []byte, error) {
		called = true
		if name != "tuist" {
			t.Fatalf("expected tuist binary, got %s", name)
		}
		expected := []string{"generate", "--no-open"}
		if len(args) != len(expected) {
			t.Fatalf("expected %d args, got %d", len(expected), len(args))
		}
		for i := range expected {
			if args[i] != expected[i] {
				t.Fatalf("arg %d: expected %q, got %q", i, expected[i], args[i])
			}
		}
		return nil, nil, nil
	}

	if err := Generate(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected command to be invoked")
	}
}

func TestGenerateNotFound(t *testing.T) {
	oldLookPath := lookPath
	t.Cleanup(func() { lookPath = oldLookPath })

	lookPath = func(string) (string, error) {
		return "", errors.New("not found")
	}

	err := Generate(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindRuntime) {
		t.Fatalf("expected runtime error kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "tuist not found in PATH") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestGenerateCommandFailureIncludesOutput(t *testing.T) {
	oldLookPath := lookPath
	oldRun := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRun
	})

	lookPath = func(string) (string, error) { return "/usr/bin/tuist", nil }
	runCommand = func(_ context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		return nil, []byte("bad generate"), errors.New("exit status 1")
	}

	err := Generate(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindRuntime) {
		t.Fatalf("expected runtime error kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "failed to generate xcode project via tuist at") || !strings.Contains(err.Error(), "bad generate") {
		t.Fatalf("expected error to include path and stderr, got %q", err.Error())
	}
}

func TestGenerateTimeout(t *testing.T) {
	oldLookPath := lookPath
	oldRun := runCommand
	oldTimeout := generateTimeout
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRun
		generateTimeout = oldTimeout
	})

	lookPath = func(string) (string, error) { return "/usr/bin/tuist", nil }
	generateTimeout = 10 * time.Millisecond
	runCommand = func(ctx context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}

	err := Generate(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !apperrors.IsKind(err, apperrors.KindRuntime) {
		t.Fatalf("expected runtime error kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "tuist generate timed out") {
		t.Fatalf("unexpected timeout message: %q", err.Error())
	}
}
