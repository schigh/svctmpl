package profile

import "io/fs"

// Profile defines the contract for a template profile.
type Profile interface {
	// Name returns the profile's identifier (e.g., "layered-http").
	Name() string
	// Description returns a human-readable description.
	Description() string
	// Manifest returns the parsed profile manifest with file conditions.
	Manifest() (*Manifest, error)
	// FS returns the template filesystem.
	// Built-in profiles use embed.FS; external profiles use os.DirFS.
	FS() fs.FS
}
