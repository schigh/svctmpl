package profile

import (
	"errors"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

// Manifest describes which files a profile generates and under what conditions.
type Manifest struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Files       []FileEntry `yaml:"files"`
}

// FileEntry describes a single template file and its output path.
type FileEntry struct {
	// Path is the template file path relative to the profile FS (e.g., "cmd/app/main.go.tmpl").
	Path string `yaml:"path"`
	// Output is the generated file path relative to the output directory (e.g., "cmd/app/main.go").
	Output string `yaml:"output"`
	// Requires is a list of conditions. All must be true (AND logic).
	// Bare string (e.g., "database") means "not none".
	// Key:value (e.g., "db_tooling:sqlc") means "equals that value".
	// Empty list means always included.
	Requires []string `yaml:"requires,omitempty"`
}

// ParseManifest parses YAML bytes into a Manifest.
// It validates that name is present, at least one file entry exists,
// and there are no duplicate output paths.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	if m.Name == "" {
		return nil, errors.New("manifest: name is required")
	}
	if len(m.Files) == 0 {
		return nil, errors.New("manifest: at least one file entry is required")
	}

	seen := make(map[string]struct{}, len(m.Files))
	for _, f := range m.Files {
		if _, ok := seen[f.Output]; ok {
			return nil, fmt.Errorf("manifest: duplicate output path %q", f.Output)
		}
		seen[f.Output] = struct{}{}
	}

	return &m, nil
}

// LoadManifest reads "profile.yaml" from the given FS and parses it.
func LoadManifest(fsys fs.FS) (*Manifest, error) {
	data, err := fs.ReadFile(fsys, "profile.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading profile.yaml: %w", err)
	}
	return ParseManifest(data)
}
