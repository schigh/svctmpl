package templates

import (
	"embed"
	"io/fs"

	"github.com/schigh/svctmpl/internal/profile"
)

//go:embed layered-http/*
var layeredHTTPFS embed.FS

// LayeredHTTP returns the built-in layered-http profile.
func LayeredHTTP() profile.Profile {
	sub, err := fs.Sub(layeredHTTPFS, "layered-http")
	if err != nil {
		panic("embedded layered-http profile is missing: " + err.Error())
	}
	return profile.NewFSProfile(
		"layered-http",
		"HTTP service with chi, postgres/sqlc, OpenTelemetry, and slog",
		sub,
	)
}

// RegisterBuiltins adds all built-in profiles to the registry.
func RegisterBuiltins(r *profile.Registry) {
	r.Register(LayeredHTTP())
}
