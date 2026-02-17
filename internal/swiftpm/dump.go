package swiftpm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

const dumpTimeout = 30 * time.Second

var lookPath = exec.LookPath

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

// DumpPackage resolves the package manifest by using swift package dump-package.
func DumpPackage(ctx context.Context, packagePath string) ([]byte, error) {
	if _, err := lookPath("swift"); err != nil {
		return nil, apperrors.New(apperrors.KindSwiftNotFound, "swift binary not found in PATH", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, dumpTimeout)
	defer cancel()

	stdout, stderr, err := runCommand(timeoutCtx, packagePath, "swift", "package", "dump-package", "--package-path", packagePath)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, apperrors.New(apperrors.KindDumpPackage, "swift package dump-package timed out", timeoutCtx.Err())
		}
		detail := strings.TrimSpace(string(stderr))
		if detail == "" {
			detail = err.Error()
		}
		return nil, apperrors.New(apperrors.KindDumpPackage, fmt.Sprintf("swift package dump-package failed: %s", detail), err)
	}
	if len(stdout) == 0 {
		return nil, apperrors.New(apperrors.KindDumpPackage, "swift package dump-package produced empty output", nil)
	}

	return stdout, nil
}
