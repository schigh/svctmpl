package profile

import (
	"fmt"
	"sort"
)

// Registry holds available profiles.
type Registry struct {
	profiles map[string]Profile
}

// NewRegistry returns an initialized Registry.
func NewRegistry() *Registry {
	return &Registry{
		profiles: make(map[string]Profile),
	}
}

// Register adds a profile to the registry.
func (r *Registry) Register(p Profile) {
	r.profiles[p.Name()] = p
}

// Get returns the profile with the given name, or an error if not found.
func (r *Registry) Get(name string) (Profile, error) {
	p, ok := r.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

// List returns all registered profiles sorted by name.
func (r *Registry) List() []Profile {
	out := make([]Profile, 0, len(r.profiles))
	for _, p := range r.profiles {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name() < out[j].Name()
	})
	return out
}
