package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/schigh/svctmpl/internal/engine"
	"github.com/schigh/svctmpl/internal/genome"
	"github.com/schigh/svctmpl/internal/templates"
)

func TestSmokeRenderAndBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping smoke test in short mode")
	}

	// Check that go toolchain is available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	g := genome.Default()
	g.Project.Name = "smoketest"
	g.Project.Module = "github.com/test/smoketest"

	p := templates.LayeredHTTP()

	outputDir := filepath.Join(t.TempDir(), "smoketest")

	eng := engine.New()
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify key files exist
	for _, f := range []string{
		"cmd/app/main.go",
		"internal/app/app.go",
		"internal/config/config.go",
		"internal/errs/errors.go",
		"internal/model/resource.go",
		"internal/service/resource.go",
		"internal/transport/http/handler.go",
		"internal/transport/http/health.go",
		"internal/repository/resource.go",
		"go.mod",
		"Makefile",
		"README.md",
	} {
		path := filepath.Join(outputDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", f)
		}
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outputDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}

	// Run go build
	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = outputDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	// Run go vet
	cmd = exec.Command("go", "vet", "./...")
	cmd.Dir = outputDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go vet failed: %v\n%s", err, out)
	}

	t.Log("Smoke test passed: rendered service compiles and vets clean")
}

func TestSmokeRenderNoDatabaseFileSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping smoke test in short mode")
	}

	g := genome.Default()
	g.Project.Name = "nodbsmoke"
	g.Project.Module = "github.com/test/nodbsmoke"
	g.Choices.Database = "none"
	g.Choices.DBTooling = "none"
	g.Choices.Migrations = "none"

	p := templates.LayeredHTTP()

	outputDir := filepath.Join(t.TempDir(), "nodbsmoke")

	eng := engine.New()
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify key files that should always exist
	for _, f := range []string{
		"cmd/app/main.go",
		"internal/app/app.go",
		"internal/config/config.go",
		"internal/errs/errors.go",
		"internal/service/resource.go",
		"internal/transport/http/handler.go",
		"internal/transport/http/health.go",
		"go.mod",
		"Makefile",
		"README.md",
	} {
		path := filepath.Join(outputDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", f)
		}
	}

	// Verify database-dependent files do NOT exist
	for _, f := range []string{
		"internal/repository/resource.go",
		"internal/repository/db.go",
		"internal/repository/models.go",
		"internal/repository/queries.sql.go",
		"internal/repository/queries.sql",
		"migrations/001_create_resources.sql",
		"sqlc.yaml",
		"internal/app/database.go",
	} {
		path := filepath.Join(outputDir, f)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("file %s should NOT exist when database=none", f)
		}
	}

	// Verify the generated output compiles
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outputDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}

	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = outputDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	t.Log("No-database smoke test passed: correct files, compiles clean")
}
