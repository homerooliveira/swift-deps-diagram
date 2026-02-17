package apperrors

import (
	"errors"
	"testing"
)

func TestErrorToExitCodeMapping(t *testing.T) {
	if code := ExitCode(New(KindInvalidArgs, "bad", nil)); code != 1 {
		t.Fatalf("expected code 1 for invalid args, got %d", code)
	}
	if code := ExitCode(New(KindManifestNotFound, "missing", nil)); code != 1 {
		t.Fatalf("expected code 1 for manifest not found, got %d", code)
	}
	if code := ExitCode(New(KindInputNotFound, "missing input", nil)); code != 1 {
		t.Fatalf("expected code 1 for input not found, got %d", code)
	}
	if code := ExitCode(New(KindXcodeProjectNotFound, "missing xcode project", nil)); code != 1 {
		t.Fatalf("expected code 1 for xcode project not found, got %d", code)
	}
	if code := ExitCode(New(KindRuntime, "boom", errors.New("x"))); code != 2 {
		t.Fatalf("expected code 2 for runtime, got %d", code)
	}
}
