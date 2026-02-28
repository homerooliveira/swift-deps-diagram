package tuist

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

var lookPath = exec.LookPath

var generateTimeout = 2 * time.Minute

var runCommand = func(ctx context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// Generate runs `tuist generate --no-open` in a directory containing Project.swift.
func Generate(ctx context.Context, path string) error {
	if _, err := lookPath("tuist"); err != nil {
		return apperrors.New(apperrors.KindRuntime, "tuist not found in PATH", err)
	}

	generateCtx, cancel := context.WithTimeout(ctx, generateTimeout)
	defer cancel()

	_, stderr, err := runCommand(generateCtx, path, "tuist", "generate", "--no-open")
	if err == nil {
		return nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return apperrors.New(apperrors.KindRuntime, "tuist not found in PATH", err)
	}
	if generateCtx.Err() == context.DeadlineExceeded {
		return apperrors.New(apperrors.KindRuntime, "tuist generate timed out", generateCtx.Err())
	}
	msg := strings.TrimSpace(string(stderr))
	if msg == "" {
		msg = err.Error()
	}
	return apperrors.New(apperrors.KindRuntime, fmt.Sprintf("failed to generate xcode project via tuist at %s: %s", path, msg), err)
}
