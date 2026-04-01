package verify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_SuccessfulProject(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal valid Go module.
	goMod := `module example.com/testproject

go 1.22.0
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

func main() {}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := Run(dir)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	for _, r := range results {
		if !r.Passed {
			t.Errorf("step %q failed: %s", r.Step, r.Output)
		}
	}
}

func TestRun_BuildFailure(t *testing.T) {
	dir := t.TempDir()

	// Write a Go module with invalid code.
	goMod := `module example.com/badproject

go 1.22.0
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	badGo := `package main

func main() {
	// This will not compile.
	x :=
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(badGo), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := Run(dir)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	// Find the build step.
	var buildResult *Result
	for i := range results {
		if results[i].Step == "go build ./..." {
			buildResult = &results[i]
			break
		}
	}

	if buildResult == nil {
		t.Fatal("expected a 'go build ./...' result")
	}

	if buildResult.Passed {
		t.Error("expected build step to fail for bad code")
	}

	if buildResult.Output == "" {
		t.Error("expected build failure output to be non-empty")
	}
}
