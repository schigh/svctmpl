package verify

import (
	"fmt"
	"os/exec"
)

// Result holds the outcome of a verification step.
type Result struct {
	Step   string
	Passed bool
	Output string
}

// Run executes post-generation verification on the generated project.
// It runs: go mod tidy, go build ./..., go vet ./..., go test ./...
// If the Go toolchain is not available, returns a single skipped result.
func Run(dir string) ([]Result, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return []Result{
			{
				Step:   "go toolchain",
				Passed: false,
				Output: "Go toolchain not found in PATH; skipping verification",
			},
		}, nil
	}

	steps := []struct {
		name string
		args []string
	}{
		{"go mod tidy", []string{goPath, "mod", "tidy"}},
		{"go build ./...", []string{goPath, "build", "./..."}},
		{"go vet ./...", []string{goPath, "vet", "./..."}},
		{"go test ./...", []string{goPath, "test", "./..."}},
	}

	var results []Result
	for _, s := range steps {
		cmd := exec.Command(s.args[0], s.args[1:]...)
		cmd.Dir = dir

		output, err := cmd.CombinedOutput()
		r := Result{
			Step:   s.name,
			Output: string(output),
			Passed: err == nil,
		}
		if err != nil {
			r.Output = fmt.Sprintf("%s\n%s", r.Output, err.Error())
		}
		results = append(results, r)
	}

	return results, nil
}
