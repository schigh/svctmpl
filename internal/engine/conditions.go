package engine

import (
	"fmt"
	"strings"

	"github.com/schigh/svctmpl/internal/genome"
)

// choicesMap returns a flat map of choice axis names to their values.
// Boolean axes are represented as "true"/"false" strings so they work
// with the same bare-key condition logic (bare "compose" passes when
// the value is not "none" and not "false").
func choicesMap(c *genome.Choices) map[string]string {
	return map[string]string{
		"transport":     c.Transport,
		"router":        c.Router,
		"database":      c.Database,
		"db_tooling":    c.DBTooling,
		"migrations":    c.Migrations,
		"structure":     c.Structure,
		"observability": c.Observability,
		"logging":       c.Logging,
		"config":        c.Config,
		"ci":            c.CI,
		"container":     c.Container,
		"compose":       fmt.Sprintf("%t", c.Compose),
		"k8s":           fmt.Sprintf("%t", c.K8s),
		"tilt":          fmt.Sprintf("%t", c.Tilt),
	}
}

// EvaluateConditions checks whether a file should be included based on genome choices.
// Each condition in the requires list must be true (AND logic).
// A bare string like "database" means the corresponding choice is not "none".
// A "key:value" string like "db_tooling:sqlc" means the choice equals that value.
// An empty requires list means always included.
func EvaluateConditions(requires []string, choices *genome.Choices) bool {
	if len(requires) == 0 {
		return true
	}

	m := choicesMap(choices)

	for _, req := range requires {
		if key, val, ok := strings.Cut(req, ":"); ok {
			// key:value form — choice must equal exact value.
			v, exists := m[key]
			if !exists || v != val {
				return false
			}
		} else {
			// bare key — choice must be active (not "none" and not "false").
			v, exists := m[req]
			if !exists || v == "none" || v == "false" {
				return false
			}
		}
	}

	return true
}
