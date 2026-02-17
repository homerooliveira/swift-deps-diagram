package output

import (
	"bytes"
	"os"
	"path/filepath"
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
