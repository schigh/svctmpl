package genome

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// validGenomeYAML returns a minimal valid genome YAML string.
func validGenomeYAML() string {
	return `version: "1"
project:
  name: myservice
  module: github.com/acme/myservice
choices:
  transport: http
  router: chi
  database: postgres
  db_tooling: sqlc
  migrations: goose
  structure: layered
  observability: otel-full
  logging: slog
  config: env
  ci: github-actions
  container: dockerfile
`
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "genome.yaml", validGenomeYAML())

	g, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Project.Name != "myservice" {
		t.Errorf("expected name myservice, got %s", g.Project.Name)
	}
	if g.Choices.Router != "chi" {
		t.Errorf("expected router chi, got %s", g.Choices.Router)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/tmp/does-not-exist-genome.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	var target *ErrFileNotFound
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrFileNotFound, got %T: %v", err, err)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "bad.yaml", ":\n  :\n  - :\n[invalid yaml{{{")

	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	var target *ErrParseFailed
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrParseFailed, got %T: %v", err, err)
	}
}

func TestLoadMissingName(t *testing.T) {
	src := `version: "1"
project:
  name: ""
  module: github.com/acme/foo
choices:
  transport: http
  router: chi
  database: postgres
  db_tooling: sqlc
  migrations: goose
  structure: layered
  observability: otel-full
  logging: slog
  config: env
  ci: github-actions
  container: dockerfile
`
	dir := t.TempDir()
	p := writeFile(t, dir, "genome.yaml", src)

	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestLoadInvalidRouter(t *testing.T) {
	src := `version: "1"
project:
  name: myservice
  module: github.com/acme/myservice
choices:
  transport: http
  router: gorilla
  database: postgres
  db_tooling: sqlc
  migrations: goose
  structure: layered
  observability: otel-full
  logging: slog
  config: env
  ci: github-actions
  container: dockerfile
`
	dir := t.TempDir()
	p := writeFile(t, dir, "genome.yaml", src)

	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for invalid router")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestCrossAxisDatabaseNoneWithTooling(t *testing.T) {
	src := `version: "1"
project:
  name: myservice
  module: github.com/acme/myservice
choices:
  transport: http
  router: chi
  database: none
  db_tooling: sqlc
  migrations: none
  structure: layered
  observability: otel-full
  logging: slog
  config: env
  ci: github-actions
  container: dockerfile
`
	dir := t.TempDir()
	p := writeFile(t, dir, "genome.yaml", src)

	_, err := Load(p)
	if err == nil {
		t.Fatal("expected validation error for db_tooling with database=none")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestCrossAxisDatabaseNoneAllNone(t *testing.T) {
	src := `version: "1"
project:
  name: myservice
  module: github.com/acme/myservice
choices:
  transport: http
  router: chi
  database: none
  db_tooling: none
  migrations: none
  structure: layered
  observability: otel-full
  logging: slog
  config: env
  ci: github-actions
  container: dockerfile
`
	dir := t.TempDir()
	p := writeFile(t, dir, "genome.yaml", src)

	g, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Choices.Database != "none" {
		t.Errorf("expected database none, got %s", g.Choices.Database)
	}
}

func TestProjectNameWithHyphens(t *testing.T) {
	g := Default()
	g.Project.Name = "my-service"
	g.Project.Module = "github.com/acme/myservice"

	err := g.Validate()
	if err == nil {
		t.Fatal("expected validation error for hyphenated name")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestProjectNameIsGoKeyword(t *testing.T) {
	g := Default()
	g.Project.Name = "type"
	g.Project.Module = "github.com/acme/myservice"

	err := g.Validate()
	if err == nil {
		t.Fatal("expected validation error for Go keyword name")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestModulePathWithoutDomain(t *testing.T) {
	g := Default()
	g.Project.Name = "myservice"
	g.Project.Module = "myservice"

	err := g.Validate()
	if err == nil {
		t.Fatal("expected validation error for module without domain")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "genome.yaml")

	original := Default()
	original.Project.Name = "roundtrip"
	original.Project.Module = "github.com/acme/roundtrip"

	if err := original.Save(p); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(p)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Compare by re-marshaling both.
	origData, _ := yaml.Marshal(original)
	loadedData, _ := yaml.Marshal(loaded)
	if string(origData) != string(loadedData) {
		t.Errorf("round-trip mismatch:\noriginal:\n%s\nloaded:\n%s", origData, loadedData)
	}
}

func TestDefaultPassesValidate(t *testing.T) {
	g := Default()
	g.Project.Name = "validservice"
	g.Project.Module = "github.com/acme/validservice"

	if err := g.Validate(); err != nil {
		t.Fatalf("Default() genome should pass validation: %v", err)
	}
}

func TestBooleanFieldsYAMLRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "genome.yaml")

	g := Default()
	g.Project.Name = "booltest"
	g.Project.Module = "github.com/acme/booltest"
	g.Choices.Compose = true
	g.Choices.K8s = false
	g.Choices.Tilt = true

	if err := g.Save(p); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(p)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Choices.Compose != true {
		t.Error("expected Compose=true after round-trip")
	}
	if loaded.Choices.K8s != false {
		t.Error("expected K8s=false after round-trip")
	}
	if loaded.Choices.Tilt != true {
		t.Error("expected Tilt=true after round-trip")
	}
}

func TestProjectNameWithBrackets(t *testing.T) {
	g := Default()
	g.Project.Name = "bad[[name"
	g.Project.Module = "github.com/acme/badname"

	err := g.Validate()
	if err == nil {
		t.Fatal("expected validation error for name with brackets")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestModulePathWithBrackets(t *testing.T) {
	g := Default()
	g.Project.Name = "myservice"
	g.Project.Module = "github.com/acme/my[[service"

	err := g.Validate()
	if err == nil {
		t.Fatal("expected validation error for module with brackets")
	}
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatalf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestAllAxesReturnsExpectedAxes(t *testing.T) {
	axes := AllAxes()
	expected := map[string]bool{
		"transport": true, "router": true, "database": true,
		"db_tooling": true, "migrations": true, "structure": true,
		"observability": true, "logging": true, "config": true,
		"ci": true, "container": true,
	}
	if len(axes) != len(expected) {
		t.Fatalf("expected %d axes, got %d", len(expected), len(axes))
	}
	for _, a := range axes {
		if !expected[a] {
			t.Errorf("unexpected axis: %s", a)
		}
	}
}

func TestAllAxesReturnsCopy(t *testing.T) {
	a1 := AllAxes()
	a2 := AllAxes()
	a1[0] = "mutated"
	if a2[0] == "mutated" {
		t.Error("AllAxes should return a copy, not the original slice")
	}
}

func TestAllowedValuesRouter(t *testing.T) {
	vals := AllowedValues("router")
	if len(vals) != 2 {
		t.Fatalf("expected 2 router values, got %d", len(vals))
	}
	expected := map[string]bool{"chi": true, "stdlib": true}
	for _, v := range vals {
		if !expected[v] {
			t.Errorf("unexpected router value: %s", v)
		}
	}
}

func TestAllowedValuesUnknownAxis(t *testing.T) {
	vals := AllowedValues("nonexistent")
	if vals != nil {
		t.Errorf("expected nil for unknown axis, got %v", vals)
	}
}

func TestIsImplemented(t *testing.T) {
	if !IsImplemented("router", "chi") {
		t.Error("expected chi to be implemented")
	}
	if IsImplemented("router", "stdlib") {
		t.Error("expected stdlib to not be implemented")
	}
	if IsImplemented("unknown", "value") {
		t.Error("expected unknown axis to return false")
	}
}
