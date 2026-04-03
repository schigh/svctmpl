package cli

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/schigh/svctmpl/internal/genome"
)

// ErrUserCancelled is returned when the user cancels the wizard.
var ErrUserCancelled = errors.New("user cancelled")

// axisDescriptions provides human-readable descriptions for select options.
var axisDescriptions = map[string]map[string]string{
	"router": {
		"chi":    "lightweight, idiomatic, composable",
		"stdlib": "standard library net/http mux (Go 1.22+)",
	},
	"database": {
		"postgres": "PostgreSQL via pgx driver",
		"sqlite":   "SQLite via modernc.org/sqlite",
		"none":     "no database",
	},
	"db_tooling": {
		"sqlc": "type-safe SQL code generation",
		"ent":  "Go-first entity framework with codegen",
		"none": "no DB tooling",
	},
	"migrations": {
		"goose":          "versioned SQL migration tool",
		"golang-migrate": "schema migration via golang-migrate",
		"atlas":          "declarative schema migrations (pairs with Ent)",
		"none":           "no migrations",
	},
	"observability": {
		"otel-full":        "OpenTelemetry traces + metrics",
		"otel-traces-only": "OpenTelemetry traces only",
		"none":             "no observability instrumentation",
	},
	"logging": {
		"slog": "standard library structured logging",
		"zap":  "Uber's high-performance logger",
	},
	"config": {
		"env":   "environment variables (os.Getenv)",
		"koanf": "koanf multi-source config library",
	},
	"ci": {
		"github-actions": "GitHub Actions CI workflow",
		"none":           "no CI configuration",
	},
	"container": {
		"dockerfile": "multi-stage Dockerfile",
		"none":       "no container configuration",
	},
}

// selectOptions builds huh.Option slice for the given axis.
func selectOptions(axis string) []huh.Option[string] {
	values := genome.AllowedValues(axis)
	descs := axisDescriptions[axis]
	opts := make([]huh.Option[string], len(values))
	for i, v := range values {
		label := v
		if d, ok := descs[v]; ok {
			label = v + " — " + d
		}
		opts[i] = huh.NewOption(label, v)
	}
	return opts
}

// RunWizard runs the interactive TUI wizard to build a Genome.
// It pre-fills any values already set in the provided genome (from CLI flags).
// Returns an error if the user cancels.
func RunWizard(g *genome.Genome) error {
	// Ensure defaults are set for any empty choices so selects have a value.
	defaults := genome.Default()
	if g.Choices.Router == "" {
		g.Choices.Router = defaults.Choices.Router
	}
	if g.Choices.Database == "" {
		g.Choices.Database = defaults.Choices.Database
	}
	if g.Choices.DBTooling == "" {
		g.Choices.DBTooling = defaults.Choices.DBTooling
	}
	if g.Choices.Migrations == "" {
		g.Choices.Migrations = defaults.Choices.Migrations
	}
	if g.Choices.Observability == "" {
		g.Choices.Observability = defaults.Choices.Observability
	}
	if g.Choices.Logging == "" {
		g.Choices.Logging = defaults.Choices.Logging
	}
	if g.Choices.Config == "" {
		g.Choices.Config = defaults.Choices.Config
	}
	if g.Choices.CI == "" {
		g.Choices.CI = defaults.Choices.CI
	}
	if g.Choices.Container == "" {
		g.Choices.Container = defaults.Choices.Container
	}

	// Pre-populate multi-select from existing boolean flags.
	var deployChoices []string
	if g.Choices.Compose {
		deployChoices = append(deployChoices, "compose")
	}
	if g.Choices.K8s {
		deployChoices = append(deployChoices, "k8s")
	}
	if g.Choices.Tilt {
		deployChoices = append(deployChoices, "tilt")
	}

	form := huh.NewForm(
		// Screen 1: Project name + module path
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Description("A valid Go identifier (lowercase, no hyphens)").
				Value(&g.Project.Name).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("project name is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Module Path").
				Description("Go module path (e.g., github.com/you/myservice)").
				Value(&g.Project.Module).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("module path is required")
					}
					return nil
				}),
		),

		// Screen 2: Router
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("HTTP Router").
				Description("Choose the HTTP router for your service").
				Options(selectOptions("router")...).
				Value(&g.Choices.Router),
		),

		// Screen 3: Database + DB tooling
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Database").
				Description("Choose the database backend").
				Options(selectOptions("database")...).
				Value(&g.Choices.Database),
			huh.NewSelect[string]().
				Title("DB Tooling").
				Description("Choose the database tooling layer").
				Options(selectOptions("db_tooling")...).
				Value(&g.Choices.DBTooling),
		),

		// Screen 4: Observability
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Observability").
				Description("Choose the observability stack").
				Options(selectOptions("observability")...).
				Value(&g.Choices.Observability),
		),

		// Screen 5: Logging + Config
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Logging").
				Description("Choose the logging library").
				Options(selectOptions("logging")...).
				Value(&g.Choices.Logging),
			huh.NewSelect[string]().
				Title("Configuration").
				Description("Choose the config loading strategy").
				Options(selectOptions("config")...).
				Value(&g.Choices.Config),
		),

		// Screen 6: CI + Container
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("CI Pipeline").
				Description("Choose the CI/CD configuration").
				Options(selectOptions("ci")...).
				Value(&g.Choices.CI),
			huh.NewSelect[string]().
				Title("Container").
				Description("Choose the container configuration").
				Options(selectOptions("container")...).
				Value(&g.Choices.Container),
		),

		// Screen 7: Deployment artifacts (multi-select)
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Deployment Artifacts").
				Description("Select which deployment/orchestration files to generate (space to toggle)").
				Options(
					huh.NewOption("Docker Compose — local dev stack with postgres + OTel", "compose"),
					huh.NewOption("Kubernetes — Deployment, Service, ConfigMap, Ingress, HPA", "k8s"),
					huh.NewOption("Tilt — live-reload dev loop for containers", "tilt"),
				).
				Value(&deployChoices),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserCancelled
		}
		return err
	}

	// Fill in fixed choices that have only one option.
	g.Version = "1"
	g.Choices.Transport = "http"
	g.Choices.Structure = "layered"

	// Map deployment multi-select back to boolean fields.
	g.Choices.Compose = contains(deployChoices, "compose")
	g.Choices.K8s = contains(deployChoices, "k8s")
	g.Choices.Tilt = contains(deployChoices, "tilt")

	// Auto-fix cross-axis constraint: database=none implies db_tooling=none and migrations=none.
	if g.Choices.Database == "none" {
		g.Choices.DBTooling = "none"
		g.Choices.Migrations = "none"
	}

	return nil
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
