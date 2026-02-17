package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	apperrors "swift-deps-diagram/internal/errors"
)

// Write emits content to stdout when outputPath is empty, otherwise atomically writes to a file.
func Write(content, outputPath string, stdout io.Writer) error {
	if outputPath == "" {
		if _, err := io.WriteString(stdout, content); err != nil {
			return apperrors.New(apperrors.KindOutputWrite, "failed writing output to stdout", err)
		}
		return nil
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return apperrors.New(apperrors.KindOutputWrite, fmt.Sprintf("failed creating output directory %s", dir), err)
	}

	tmpFile, err := os.CreateTemp(dir, ".swift-deps-diagram-*")
	if err != nil {
		return apperrors.New(apperrors.KindOutputWrite, "failed creating temp output file", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := io.WriteString(tmpFile, content); err != nil {
		_ = tmpFile.Close()
		return apperrors.New(apperrors.KindOutputWrite, "failed writing output file", err)
	}
	if err := tmpFile.Close(); err != nil {
		return apperrors.New(apperrors.KindOutputWrite, "failed closing output file", err)
	}

	if err := os.Rename(tmpName, outputPath); err != nil {
		return apperrors.New(apperrors.KindOutputWrite, "failed moving output file into place", err)
	}

	return nil
}
