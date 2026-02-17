package swiftpm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

func TestDumpPackageSuccess(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	lookPath = func(string) (string, error) { return "/usr/bin/swift", nil }
	runCommand = func(_ context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		return []byte(`{"name":"X","targets":[]}`), nil, nil
	}
	defer func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	}()

	out, err := DumpPackage(context.Background(), ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestDumpPackageSwiftNotFound(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		return nil, nil, nil
	}
	defer func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	}()

	_, err := DumpPackage(context.Background(), ".")
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindSwiftNotFound) {
		t.Fatalf("expected swift not found kind, got %v", err)
	}
}

func TestDumpPackageCommandFailureIncludesStderr(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	lookPath = func(string) (string, error) { return "/usr/bin/swift", nil }
	runCommand = func(_ context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		return nil, []byte("manifest error"), errors.New("exit 1")
	}
	defer func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	}()

	_, err := DumpPackage(context.Background(), ".")
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindDumpPackage) {
		t.Fatalf("expected dump package kind, got %v", err)
	}
	if msg := err.Error(); msg == "" || !strings.Contains(msg, "manifest error") {
		t.Fatalf("expected stderr detail in message, got %q", msg)
	}
}

func TestDumpPackageRespectsTimeout(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	lookPath = func(string) (string, error) { return "/usr/bin/swift", nil }
	runCommand = func(ctx context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}
	defer func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := DumpPackage(ctx, ".")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !apperrors.IsKind(err, apperrors.KindDumpPackage) {
		t.Fatalf("expected dump package kind, got %v", err)
	}
}
