package genome

import (
	"go/token"
	"strings"
	"unicode"
)

// axisOrder defines the display order for all axes.
var axisOrder = []string{
	"transport",
	"router",
	"database",
	"db_tooling",
	"migrations",
	"structure",
	"observability",
	"logging",
	"config",
	"ci",
	"container",
}

// allowedValues maps each axis to its permitted values.
var allowedValues = map[string][]string{
	"transport":     {"http"},
	"router":        {"chi", "stdlib"},
	"database":      {"postgres", "sqlite", "none"},
	"db_tooling":    {"sqlc", "none"},
	"migrations":    {"goose", "golang-migrate", "none"},
	"structure":     {"layered"},
	"observability": {"otel-full", "otel-traces-only", "none"},
	"logging":       {"slog", "zap"},
	"config":        {"env", "koanf"},
	"ci":            {"github-actions", "none"},
	"container":     {"dockerfile", "none"},
}

// implementedCombinations lists axis:value pairs that have templates in v1.
var implementedCombinations = map[string]map[string]bool{
	"router":        {"chi": true},
	"database":      {"postgres": true},
	"db_tooling":    {"sqlc": true},
	"migrations":    {"goose": true},
	"observability": {"otel-full": true},
	"logging":       {"slog": true},
	"config":        {"env": true},
}

// AllAxes returns all axis names in display order.
func AllAxes() []string {
	out := make([]string, len(axisOrder))
	copy(out, axisOrder)
	return out
}

// AllowedValues returns the permitted values for the given axis name.
// Returns nil if the axis is unknown.
func AllowedValues(axis string) []string {
	vals, ok := allowedValues[axis]
	if !ok {
		return nil
	}
	out := make([]string, len(vals))
	copy(out, vals)
	return out
}

// IsImplemented returns true if the given axis:value combination has a
// template in v1.
func IsImplemented(axis, value string) bool {
	vals, ok := implementedCombinations[axis]
	if !ok {
		return false
	}
	return vals[value]
}

// Validate checks the genome for structural correctness. It returns an
// *ErrValidation containing all violations, or nil if the genome is valid.
func (g *Genome) Validate() error {
	var msgs []string

	// Version check.
	if g.Version != "1" {
		msgs = append(msgs, `version must be "1"`)
	}

	// Project name: required, valid Go identifier, not a keyword.
	if g.Project.Name == "" {
		msgs = append(msgs, "project.name is required")
	} else {
		if !isValidGoIdentifier(g.Project.Name) {
			msgs = append(msgs, "project.name must be a valid Go identifier (lowercase letters, digits, underscores; no hyphens or spaces)")
		}
		if token.IsKeyword(g.Project.Name) {
			msgs = append(msgs, "project.name must not be a Go keyword")
		}
		if containsBrackets(g.Project.Name) {
			msgs = append(msgs, "project.name must not contain '[' or ']' characters")
		}
	}

	// Project module: required, must contain a dot.
	if g.Project.Module == "" {
		msgs = append(msgs, "project.module is required")
	} else {
		if !strings.Contains(g.Project.Module, ".") {
			msgs = append(msgs, "project.module must contain a domain (at least one '.')")
		}
		if containsBrackets(g.Project.Module) {
			msgs = append(msgs, "project.module must not contain '[' or ']' characters")
		}
	}

	// Validate each choice axis.
	choiceMap := map[string]string{
		"transport":     g.Choices.Transport,
		"router":        g.Choices.Router,
		"database":      g.Choices.Database,
		"db_tooling":    g.Choices.DBTooling,
		"migrations":    g.Choices.Migrations,
		"structure":     g.Choices.Structure,
		"observability": g.Choices.Observability,
		"logging":       g.Choices.Logging,
		"config":        g.Choices.Config,
		"ci":            g.Choices.CI,
		"container":     g.Choices.Container,
	}

	for _, axis := range axisOrder {
		val := choiceMap[axis]
		allowed := allowedValues[axis]
		if !contains(allowed, val) {
			msgs = append(msgs, axis+": invalid value "+quote(val)+"; allowed: "+strings.Join(allowed, ", "))
		}
	}

	// Cross-axis: database=none → db_tooling and migrations must be none.
	if g.Choices.Database == "none" {
		if g.Choices.DBTooling != "none" {
			msgs = append(msgs, "db_tooling must be \"none\" when database is \"none\"")
		}
		if g.Choices.Migrations != "none" {
			msgs = append(msgs, "migrations must be \"none\" when database is \"none\"")
		}
	}

	if len(msgs) > 0 {
		return &ErrValidation{Messages: msgs}
	}
	return nil
}

// isValidGoIdentifier checks that s is a non-empty identifier consisting of
// lowercase ASCII letters, digits, and underscores, starting with a letter.
func isValidGoIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLower(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

func containsBrackets(s string) bool {
	return strings.ContainsAny(s, "[]")
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func quote(s string) string {
	return `"` + s + `"`
}
