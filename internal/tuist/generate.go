package tuist

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
)

var execCommandContext = exec.CommandContext

// Generate runs `tuist generate --no-open` in a directory containing Project.swift.
func Generate(ctx context.Context, path string) error {
	cmd := execCommandContext(ctx, "tuist", "generate", "--no-open")
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
