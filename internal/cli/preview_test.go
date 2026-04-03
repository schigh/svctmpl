package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn and returns whatever it printed to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	return buf.String()
}

func TestPrintTree_SingleFile(t *testing.T) {
	out := captureStdout(t, func() {
		PrintTree("/tmp/myproject", []string{"main.go"})
	})

	if !strings.Contains(out, "myproject/") {
		t.Errorf("expected root dir in output, got:\n%s", out)
	}
	if !strings.Contains(out, "└── main.go") {
		t.Errorf("expected └── main.go in output, got:\n%s", out)
	}
}

func TestPrintTree_NestedDirectories(t *testing.T) {
	files := []string{
		"cmd/app/main.go",
		"internal/server/server.go",
		"internal/server/routes.go",
		"go.mod",
	}

	out := captureStdout(t, func() {
		PrintTree("/tmp/myproject", files)
	})

	// Verify root
	if !strings.Contains(out, "myproject/") {
		t.Errorf("expected root dir in output, got:\n%s", out)
	}

	// Verify tree connectors appear (├── for non-last, └── for last)
	if !strings.Contains(out, "├──") {
		t.Errorf("expected ├── connector in output, got:\n%s", out)
	}
	if !strings.Contains(out, "└──") {
		t.Errorf("expected └── connector in output, got:\n%s", out)
	}

	// Verify directory suffixes
	if !strings.Contains(out, "cmd/") {
		t.Errorf("expected cmd/ directory suffix in output, got:\n%s", out)
	}

	// Verify files appear
	if !strings.Contains(out, "main.go") {
		t.Errorf("expected main.go in output, got:\n%s", out)
	}
	if !strings.Contains(out, "server.go") {
		t.Errorf("expected server.go in output, got:\n%s", out)
	}
	if !strings.Contains(out, "go.mod") {
		t.Errorf("expected go.mod in output, got:\n%s", out)
	}
}

func TestPrintTree_EmptyList(t *testing.T) {
	// Should not panic on empty input.
	out := captureStdout(t, func() {
		PrintTree("/tmp/myproject", nil)
	})
	if out != "" {
		t.Errorf("expected no output for empty list, got:\n%s", out)
	}

	out = captureStdout(t, func() {
		PrintTree("/tmp/myproject", []string{})
	})
	if out != "" {
		t.Errorf("expected no output for empty slice, got:\n%s", out)
	}
}
