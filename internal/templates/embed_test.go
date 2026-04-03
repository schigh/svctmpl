package templates

import (
	"testing"

	"github.com/schigh/svctmpl/internal/profile"
)

func TestLayeredHTTP_ReturnsValidProfile(t *testing.T) {
	p := LayeredHTTP()
	if p == nil {
		t.Fatal("LayeredHTTP() returned nil")
	}
}

func TestLayeredHTTP_ProfileName(t *testing.T) {
	p := LayeredHTTP()
	if p.Name() != "layered-http" {
		t.Errorf("expected profile name %q, got %q", "layered-http", p.Name())
	}
}

func TestLayeredHTTP_ManifestHasFiles(t *testing.T) {
	p := LayeredHTTP()
	m, err := p.Manifest()
	if err != nil {
		t.Fatalf("Manifest() error: %v", err)
	}
	if len(m.Files) == 0 {
		t.Error("expected manifest to have >0 files")
	}
}

func TestLayeredHTTP_ManifestNameMatches(t *testing.T) {
	p := LayeredHTTP()
	m, err := p.Manifest()
	if err != nil {
		t.Fatalf("Manifest() error: %v", err)
	}
	if m.Name != "layered-http" {
		t.Errorf("expected manifest name %q, got %q", "layered-http", m.Name)
	}
}

func TestRegisterBuiltins_AddsProfile(t *testing.T) {
	reg := profile.NewRegistry()
	RegisterBuiltins(reg)

	p, err := reg.Get("layered-http")
	if err != nil {
		t.Fatalf("expected layered-http in registry: %v", err)
	}
	if p.Name() != "layered-http" {
		t.Errorf("expected profile name %q, got %q", "layered-http", p.Name())
	}
}

func TestRegisterBuiltins_ListNotEmpty(t *testing.T) {
	reg := profile.NewRegistry()
	RegisterBuiltins(reg)

	list := reg.List()
	if len(list) == 0 {
		t.Error("expected at least one profile after RegisterBuiltins")
	}
}
