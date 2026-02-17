package graphviz

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

const renderTimeout = 30 * time.Second

var lookPath = exec.LookPath

var runDot = func(ctx context.Context, dotSource, outputPath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "dot", "-Tpng", "-o", outputPath)
	cmd.Stdin = strings.NewReader(dotSource)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.Bytes(), err
}

// WritePNG renders dot source into a PNG file using Graphviz.
func WritePNG(ctx context.Context, dotSource, outputPath string) error {
	if outputPath == "" {
		return nil
	}
	if _, err := lookPath("dot"); err != nil {
		return apperrors.New(apperrors.KindGraphvizNotFound, "graphviz 'dot' binary not found in PATH", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return apperrors.New(apperrors.KindOutputWrite, fmt.Sprintf("failed creating png output directory %s", dir), err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, renderTimeout)
	defer cancel()

	stderr, err := runDot(timeoutCtx, dotSource, outputPath)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return apperrors.New(apperrors.KindGraphvizRender, "graphviz rendering timed out", timeoutCtx.Err())
		}
		detail := strings.TrimSpace(string(stderr))
		if detail == "" {
			detail = err.Error()
		}
		return apperrors.New(apperrors.KindGraphvizRender, fmt.Sprintf("graphviz rendering failed: %s", detail), err)
	}
	return nil
}
