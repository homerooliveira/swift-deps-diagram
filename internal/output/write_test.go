package output

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteToStdoutWhenNoOutputPath(t *testing.T) {
	var buf bytes.Buffer
	if err := Write("hello", "", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "hello" {
		t.Fatalf("unexpected stdout output %q", buf.String())
	}
}

func TestWriteToFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "diagram.txt")
	if err := Write("content", path, &bytes.Buffer{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != "content" {
		t.Fatalf("unexpected file output %q", string(data))
	}
}

func TestWriteToFileUsesReadablePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission bits are not portable on windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "diagram.txt")
	if err := Write("content", path, &bytes.Buffer{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Fatalf("expected output file mode 0644, got %#o", got)
	}
}
