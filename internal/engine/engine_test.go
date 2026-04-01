package engine

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/schigh/svctmpl/internal/genome"
	"github.com/schigh/svctmpl/internal/profile"
)

// testProfile implements profile.Profile using an in-memory filesystem.
type testProfile struct {
	name        string
	description string
	fsys        fs.FS
}

func (p *testProfile) Name() string             { return p.name }
func (p *testProfile) Description() string       { return p.description }
func (p *testProfile) FS() fs.FS                 { return p.fsys }
func (p *testProfile) Manifest() (*profile.Manifest, error) {
	return profile.LoadManifest(p.fsys)
}

// newTestGenome returns a fully-populated test genome.
func newTestGenome() *genome.Genome {
	g := genome.Default()
	g.Project.Name = "testservice"
	g.Project.Module = "github.com/acme/testservice"
	return g
}

// newTestProfile builds a profile with the given manifest YAML and template files.
func newTestProfile(manifestYAML string, files map[string]string) *testProfile {
	m := fstest.MapFS{
		"profile.yaml": &fstest.MapFile{Data: []byte(manifestYAML)},
	}
	for name, content := range files {
		m[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return &testProfile{
		name:        "test",
		description: "test profile",
		fsys:        m,
	}
}

func TestRenderHappyPath(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: main.go.tmpl
    output: cmd/app/main.go
`
	p := newTestProfile(manifest, map[string]string{
		"main.go.tmpl": `package main // [[ .ProjectName ]]`,
	})

	g := newTestGenome()
	eng := New()

	outputDir := filepath.Join(t.TempDir(), "output")
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputDir, "cmd/app/main.go"))
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	expected := "package main // testservice"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestRenderConditionIncluded(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: db.go.tmpl
    output: db.go
    requires: [database]
`
	p := newTestProfile(manifest, map[string]string{
		"db.go.tmpl": `package db // [[ .Database ]]`,
	})

	g := newTestGenome()
	g.Choices.Database = "postgres"

	eng := New()
	outputDir := filepath.Join(t.TempDir(), "output")
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "db.go")); err != nil {
		t.Errorf("expected db.go to exist: %v", err)
	}
}

func TestRenderConditionSkipped(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: db.go.tmpl
    output: db.go
    requires: [database]
`
	p := newTestProfile(manifest, map[string]string{
		"db.go.tmpl": `package db`,
	})

	g := newTestGenome()
	g.Choices.Database = "none"

	eng := New()
	outputDir := filepath.Join(t.TempDir(), "output")
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "db.go")); !os.IsNotExist(err) {
		t.Errorf("expected db.go to not exist when database=none")
	}
}

func TestRenderConditionMultipleMatch(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: sqlc.go.tmpl
    output: sqlc.go
    requires: [database, "db_tooling:sqlc"]
`
	p := newTestProfile(manifest, map[string]string{
		"sqlc.go.tmpl": `package sqlc`,
	})

	g := newTestGenome()
	g.Choices.Database = "postgres"
	g.Choices.DBTooling = "sqlc"

	eng := New()
	outputDir := filepath.Join(t.TempDir(), "output")
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "sqlc.go")); err != nil {
		t.Errorf("expected sqlc.go to exist: %v", err)
	}
}

func TestRenderConditionMultipleNoMatch(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: sqlc.go.tmpl
    output: sqlc.go
    requires: [database, "db_tooling:sqlc"]
`
	p := newTestProfile(manifest, map[string]string{
		"sqlc.go.tmpl": `package sqlc`,
	})

	g := newTestGenome()
	g.Choices.Database = "postgres"
	g.Choices.DBTooling = "none"

	eng := New()
	outputDir := filepath.Join(t.TempDir(), "output")
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "sqlc.go")); !os.IsNotExist(err) {
		t.Errorf("expected sqlc.go to not exist when db_tooling=none")
	}
}

func TestRenderAtomicOnFailure(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: bad.go.tmpl
    output: bad.go
`
	p := newTestProfile(manifest, map[string]string{
		"bad.go.tmpl": `[[ .Nonexistent.Deep.Field ]]`,
	})

	g := newTestGenome()
	eng := New()

	outputDir := filepath.Join(t.TempDir(), "output")
	err := eng.Render(g, p, outputDir)
	if err == nil {
		t.Fatal("expected render error for bad template execution")
	}

	if _, statErr := os.Stat(outputDir); !os.IsNotExist(statErr) {
		t.Errorf("expected output dir to not exist after failed render")
	}
}

func TestRenderErrorWrapping(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: bad.go.tmpl
    output: bad.go
`
	p := newTestProfile(manifest, map[string]string{
		"bad.go.tmpl": `[[ .Invalid | bad_func ]]`,
	})

	g := newTestGenome()
	eng := New()

	outputDir := filepath.Join(t.TempDir(), "output")
	err := eng.Render(g, p, outputDir)
	if err == nil {
		t.Fatal("expected error for bad template syntax")
	}

	var renderErr *RenderError
	if !errors.As(err, &renderErr) {
		t.Fatalf("expected RenderError, got %T: %v", err, err)
	}
	if renderErr.File != "bad.go.tmpl" {
		t.Errorf("expected file bad.go.tmpl, got %s", renderErr.File)
	}
}

func TestEvaluateConditionsEmptyRequires(t *testing.T) {
	choices := &genome.Choices{
		Transport: "http",
		Database:  "none",
	}
	if !EvaluateConditions(nil, choices) {
		t.Error("empty requires should return true")
	}
	if !EvaluateConditions([]string{}, choices) {
		t.Error("empty requires slice should return true")
	}
}

func TestRenderVerbose(t *testing.T) {
	manifest := `name: test
description: test profile
files:
  - path: main.go.tmpl
    output: main.go
  - path: db.go.tmpl
    output: db.go
    requires: [database]
`
	p := newTestProfile(manifest, map[string]string{
		"main.go.tmpl": `package main`,
		"db.go.tmpl":   `package db`,
	})

	g := newTestGenome()
	g.Choices.Database = "none"

	eng := New(WithVerbose(true))
	outputDir := filepath.Join(t.TempDir(), "output")
	if err := eng.Render(g, p, outputDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// main.go should exist, db.go should not.
	if _, err := os.Stat(filepath.Join(outputDir, "main.go")); err != nil {
		t.Errorf("expected main.go to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "db.go")); !os.IsNotExist(err) {
		t.Errorf("expected db.go to not exist")
	}
}
