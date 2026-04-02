// Package genome defines the decision schema for svctmpl service generation.
// A Genome is a YAML manifest capturing every structural choice needed to
// scaffold a Go service.
package genome

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Genome represents all decisions for generating a service.
type Genome struct {
	Version        string        `yaml:"version"`
	GeneratedAt    string        `yaml:"generated_at,omitempty"`
	SvctmplVersion string        `yaml:"svctmpl_version,omitempty"`
	Project        ProjectConfig `yaml:"project"`
	Choices        Choices       `yaml:"choices"`
}

// ProjectConfig holds the project name and Go module path.
type ProjectConfig struct {
	Name   string `yaml:"name"`
	Module string `yaml:"module"`
}

// Choices captures every structural axis for code generation.
type Choices struct {
	Transport     string `yaml:"transport"`
	Router        string `yaml:"router"`
	Database      string `yaml:"database"`
	DBTooling     string `yaml:"db_tooling"`
	Migrations    string `yaml:"migrations"`
	Structure     string `yaml:"structure"`
	Observability string `yaml:"observability"`
	Logging       string `yaml:"logging"`
	Config        string `yaml:"config"`
	CI            string `yaml:"ci"`
	Container     string `yaml:"container"`

	// Deployment artifacts (boolean, all optional)
	Compose bool `yaml:"compose"`
	K8s     bool `yaml:"k8s"`
	Tilt    bool `yaml:"tilt"`
}

// ErrFileNotFound indicates the genome file does not exist.
type ErrFileNotFound struct {
	Path string
	Err  error
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("genome file not found: %s", e.Path)
}

func (e *ErrFileNotFound) Unwrap() error {
	return e.Err
}

// ErrParseFailed indicates the YAML could not be parsed.
type ErrParseFailed struct {
	Path string
	Err  error
}

func (e *ErrParseFailed) Error() string {
	return fmt.Sprintf("failed to parse genome file %s: %v", e.Path, e.Err)
}

func (e *ErrParseFailed) Unwrap() error {
	return e.Err
}

// ErrValidation contains one or more validation failures.
type ErrValidation struct {
	Messages []string
}

func (e *ErrValidation) Error() string {
	if len(e.Messages) == 1 {
		return fmt.Sprintf("genome validation failed: %s", e.Messages[0])
	}
	return fmt.Sprintf("genome validation failed (%d issues): %v", len(e.Messages), e.Messages)
}

// Load reads a YAML genome file, unmarshals it, and validates the result.
// It returns distinct error types for file-not-found, parse errors, and
// validation errors.
func Load(path string) (*Genome, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &ErrFileNotFound{Path: path, Err: err}
		}
		return nil, fmt.Errorf("reading genome file: %w", err)
	}

	var g Genome
	if err := yaml.Unmarshal(data, &g); err != nil {
		return nil, &ErrParseFailed{Path: path, Err: err}
	}

	if err := g.Validate(); err != nil {
		return nil, err
	}

	return &g, nil
}

// Save marshals the genome to YAML and writes it to path.
func (g *Genome) Save(path string) error {
	data, err := yaml.Marshal(g)
	if err != nil {
		return fmt.Errorf("marshaling genome: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Default returns a Genome with all default values filled in.
func Default() *Genome {
	return &Genome{
		Version: "1",
		Choices: Choices{
			Transport:     "http",
			Router:        "chi",
			Database:      "postgres",
			DBTooling:     "sqlc",
			Migrations:    "goose",
			Structure:     "layered",
			Observability: "otel-full",
			Logging:       "slog",
			Config:        "env",
			CI:            "github-actions",
			Container:     "dockerfile",
			Compose:       true,
			K8s:           true,
			Tilt:          false,
		},
	}
}
