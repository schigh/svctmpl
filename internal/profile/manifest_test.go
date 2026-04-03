package profile

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseManifest_Valid(t *testing.T) {
	data := []byte(`
name: test-profile
description: A test profile
files:
  - path: cmd/app/main.go.tmpl
    output: cmd/app/main.go
  - path: internal/server/server.go.tmpl
    output: internal/server/server.go
`)
	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "test-profile" {
		t.Errorf("Name = %q, want %q", m.Name, "test-profile")
	}
	if m.Description != "A test profile" {
		t.Errorf("Description = %q, want %q", m.Description, "A test profile")
	}
	if len(m.Files) != 2 {
		t.Fatalf("len(Files) = %d, want 2", len(m.Files))
	}
	if m.Files[0].Path != "cmd/app/main.go.tmpl" {
		t.Errorf("Files[0].Path = %q, want %q", m.Files[0].Path, "cmd/app/main.go.tmpl")
	}
	if m.Files[0].Output != "cmd/app/main.go" {
		t.Errorf("Files[0].Output = %q, want %q", m.Files[0].Output, "cmd/app/main.go")
	}
}

func TestParseManifest_MissingName(t *testing.T) {
	data := []byte(`
description: no name
files:
  - path: foo.tmpl
    output: foo.go
`)
	_, err := ParseManifest(data)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("error = %q, want it to mention 'name is required'", err)
	}
}

func TestParseManifest_NoFiles(t *testing.T) {
	data := []byte(`
name: empty-profile
description: no files
files: []
`)
	_, err := ParseManifest(data)
	if err == nil {
		t.Fatal("expected error for no files")
	}
	if !strings.Contains(err.Error(), "at least one file entry") {
		t.Errorf("error = %q, want it to mention 'at least one file entry'", err)
	}
}

func TestParseManifest_DuplicateOutputPaths(t *testing.T) {
	data := []byte(`
name: dup-profile
files:
  - path: a.tmpl
    output: out.go
  - path: b.tmpl
    output: out.go
`)
	_, err := ParseManifest(data)
	if err == nil {
		t.Fatal("expected error for duplicate output paths")
	}
	if !strings.Contains(err.Error(), "duplicate output path") {
		t.Errorf("error = %q, want it to mention 'duplicate output path'", err)
	}
}

func TestParseManifest_RequiresConditions(t *testing.T) {
	data := []byte(`
name: cond-profile
files:
  - path: db.go.tmpl
    output: db.go
    requires:
      - database
      - db_tooling:sqlc
`)
	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Files[0].Requires) != 2 {
		t.Fatalf("len(Requires) = %d, want 2", len(m.Files[0].Requires))
	}
	if m.Files[0].Requires[0] != "database" {
		t.Errorf("Requires[0] = %q, want %q", m.Files[0].Requires[0], "database")
	}
	if m.Files[0].Requires[1] != "db_tooling:sqlc" {
		t.Errorf("Requires[1] = %q, want %q", m.Files[0].Requires[1], "db_tooling:sqlc")
	}
}

func TestParseManifest_UnknownFields(t *testing.T) {
	data := []byte(`
name: future-profile
future_field: something
files:
  - path: main.go.tmpl
    output: main.go
    unknown_attr: 42
`)
	_, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("unexpected error for unknown fields: %v", err)
	}
}

func TestRegistry_RegisterGetList(t *testing.T) {
	r := NewRegistry()

	fsys := fstest.MapFS{
		"profile.yaml": &fstest.MapFile{
			Data: []byte("name: alpha\nfiles:\n  - path: a.tmpl\n    output: a.go\n"),
		},
	}
	p1 := NewFSProfile("alpha", "First", fsys)
	p2 := NewFSProfile("beta", "Second", fsys)

	r.Register(p1)
	r.Register(p2)

	got, err := r.Get("alpha")
	if err != nil {
		t.Fatalf("Get(alpha): %v", err)
	}
	if got.Name() != "alpha" {
		t.Errorf("Get(alpha).Name() = %q", got.Name())
	}

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("List() len = %d, want 2", len(list))
	}
	if list[0].Name() != "alpha" || list[1].Name() != "beta" {
		t.Errorf("List() = [%q, %q], want [alpha, beta]", list[0].Name(), list[1].Name())
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to mention 'not found'", err)
	}
}
