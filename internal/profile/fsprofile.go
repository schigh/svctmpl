package profile

import "io/fs"

// FSProfile implements Profile using any fs.FS.
// Used for both embedded (embed.FS) and disk-based (os.DirFS) profiles.
type FSProfile struct {
	name string
	desc string
	fsys fs.FS
}

// NewFSProfile creates a new FSProfile with the given name, description, and filesystem.
func NewFSProfile(name, desc string, fsys fs.FS) *FSProfile {
	return &FSProfile{
		name: name,
		desc: desc,
		fsys: fsys,
	}
}

// Name returns the profile's identifier.
func (p *FSProfile) Name() string { return p.name }

// Description returns a human-readable description.
func (p *FSProfile) Description() string { return p.desc }

// Manifest returns the parsed profile manifest.
func (p *FSProfile) Manifest() (*Manifest, error) {
	return LoadManifest(p.fsys)
}

// FS returns the template filesystem.
func (p *FSProfile) FS() fs.FS { return p.fsys }
