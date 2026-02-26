package tuist

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
)

// Generate runs `tuist generate` in a directory containing Project.swift.
func Generate(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "tuist", "generate")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return apperrors.New(apperrors.KindRuntime, "tuist not found in PATH", err)
	}
	msg := strings.TrimSpace(string(output))
	if msg == "" {
		msg = "tuist generate failed"
	}
	return apperrors.New(apperrors.KindRuntime, fmt.Sprintf("failed to generate xcode project via tuist at %s", path), errors.New(msg))
}
