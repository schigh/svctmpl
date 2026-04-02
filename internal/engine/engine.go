package engine

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/schigh/svctmpl/internal/genome"
	"github.com/schigh/svctmpl/internal/profile"
)

// RenderError wraps a template rendering error with the source filename.
type RenderError struct {
	File string
	Err  error
}

func (e *RenderError) Error() string {
	return fmt.Sprintf("rendering %s: %v", e.File, e.Err)
}

func (e *RenderError) Unwrap() error {
	return e.Err
}

// TemplateData holds the values available inside every template.
type TemplateData struct {
	ProjectName   string
	ModulePath    string
	Transport     string
	Router        string
	Database      string
	DBTooling     string
	Migrations    string
	Structure     string
	Observability string
	Logging       string
	Config        string
	CI            string
	Container     string

	// Convenience booleans
	HasDatabase  bool // database != "none"
	HasDBTooling bool // db_tooling != "none"
	HasMigrations bool // migrations != "none"
	HasOTel      bool // observability != "none"
	HasOTelFull  bool // observability == "otel-full"
	HasCI        bool // ci != "none"
	HasContainer bool // container != "none"
	HasCompose   bool // compose == true
	HasK8s       bool // k8s == true
	HasTilt      bool // tilt == true
}

// NewTemplateData builds a TemplateData from a Genome.
func NewTemplateData(g *genome.Genome) *TemplateData {
	c := &g.Choices
	return &TemplateData{
		ProjectName:   g.Project.Name,
		ModulePath:    g.Project.Module,
		Transport:     c.Transport,
		Router:        c.Router,
		Database:      c.Database,
		DBTooling:     c.DBTooling,
		Migrations:    c.Migrations,
		Structure:     c.Structure,
		Observability: c.Observability,
		Logging:       c.Logging,
		Config:        c.Config,
		CI:            c.CI,
		Container:     c.Container,

		HasDatabase:   c.Database != "none",
		HasDBTooling:  c.DBTooling != "none",
		HasMigrations: c.Migrations != "none",
		HasOTel:       c.Observability != "none",
		HasOTelFull:   c.Observability == "otel-full",
		HasCI:         c.CI != "none",
		HasContainer:  c.Container != "none",
		HasCompose:    c.Compose,
		HasK8s:        c.K8s,
		HasTilt:       c.Tilt,
	}
}

// Engine renders templates from a profile using genome choices.
type Engine struct {
	verbose bool
}

// Option configures the engine.
type Option func(*Engine)

// WithVerbose enables verbose logging during rendering.
func WithVerbose(v bool) Option {
	return func(e *Engine) {
		e.verbose = v
	}
}

// New creates a new Engine with the given options.
func New(opts ...Option) *Engine {
	e := &Engine{}
	for _, o := range opts {
		o(e)
	}
	return e
}

// Render generates files from a profile into outputDir based on genome choices.
// It uses atomic writes: renders to a temp dir in outputDir's parent, then
// renames on success. On failure, the temp dir is cleaned up.
func (e *Engine) Render(g *genome.Genome, p profile.Profile, outputDir string) error {
	manifest, err := p.Manifest()
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	parentDir := filepath.Dir(outputDir)
	tempDir, err := os.MkdirTemp(parentDir, ".svctmpl-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}

	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(tempDir)
		}
	}()

	data := NewTemplateData(g)
	fsys := p.FS()

	for _, entry := range manifest.Files {
		if !EvaluateConditions(entry.Requires, &g.Choices) {
			if e.verbose {
				log.Printf("skipping %s: conditions not met", entry.Path)
			}
			continue
		}

		tmplBytes, err := fs.ReadFile(fsys, entry.Path)
		if err != nil {
			return &RenderError{File: entry.Path, Err: fmt.Errorf("reading template: %w", err)}
		}

		tmpl, err := template.New(entry.Path).Delims("[[", "]]").Parse(string(tmplBytes))
		if err != nil {
			return &RenderError{File: entry.Path, Err: fmt.Errorf("parsing template: %w", err)}
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return &RenderError{File: entry.Path, Err: fmt.Errorf("executing template: %w", err)}
		}

		outPath := filepath.Join(tempDir, entry.Output)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("creating directories for %s: %w", entry.Output, err)
		}

		if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", entry.Output, err)
		}

		if e.verbose {
			log.Printf("rendered %s -> %s", entry.Path, entry.Output)
		}
	}

	if err := os.Rename(tempDir, outputDir); err != nil {
		return fmt.Errorf("moving rendered output to %s: %w", outputDir, err)
	}
	success = true

	return nil
}
