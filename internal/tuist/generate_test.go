package tuist

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
)

func TestGenerateSuccessUsesNoOpenFlag(t *testing.T) {
	oldExec := execCommandContext
	t.Cleanup(func() { execCommandContext = oldExec })

	called := false
	execCommandContext = func(_ context.Context, name string, args ...string) *exec.Cmd {
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
		return exec.CommandContext(context.Background(), "echo")
	}

	if err := Generate(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected command to be invoked")
	}
}

func TestGenerateNotFound(t *testing.T) {
	oldExec := execCommandContext
	t.Cleanup(func() { execCommandContext = oldExec })

	execCommandContext = func(_ context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(context.Background(), "definitely-not-a-real-binary-for-tuist-test")
	}

	err := Generate(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindRuntime) {
		t.Fatalf("expected runtime error kind, got %v", err)
	}
}

func TestGenerateCommandFailureIncludesOutput(t *testing.T) {
	oldExec := execCommandContext
	t.Cleanup(func() { execCommandContext = oldExec })

	execCommandContext = func(_ context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(context.Background(), "sh", "-c", "echo 'bad generate' >&2; exit 1")
	}

	err := Generate(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindRuntime) {
		t.Fatalf("expected runtime error kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "bad generate") {
		t.Fatalf("expected tool output in error, got %q", err.Error())
	}
}
