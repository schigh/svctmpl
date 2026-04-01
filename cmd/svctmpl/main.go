package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/schigh/svctmpl/internal/cli"
	"github.com/schigh/svctmpl/internal/engine"
	"github.com/schigh/svctmpl/internal/genome"
	"github.com/schigh/svctmpl/internal/profile"
	"github.com/schigh/svctmpl/internal/templates"
	"github.com/schigh/svctmpl/internal/verify"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "svctmpl",
		Short:   "Scaffold production-ready Go services",
		Version: version,
	}

	cmd.AddCommand(newCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(completionCmd())

	return cmd
}

func newCmd() *cobra.Command {
	var (
		fromFile      string
		outputDir     string
		force         bool
		devProfileDir string
		verbose       bool

		flagName          string
		flagModule        string
		flagRouter        string
		flagDatabase      string
		flagDBTooling     string
		flagMigrations    string
		flagObservability string
		flagLogging       string
		flagConfig        string
		flagCI            string
		flagContainer     string
	)

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new Go service from a template",
		Long:  "Create a new Go service interactively or from a genome YAML file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var g *genome.Genome

			if fromFile != "" {
				// Load genome from file, skip TUI.
				loaded, err := genome.Load(fromFile)
				if err != nil {
					return fmt.Errorf("loading genome file: %w", err)
				}
				g = loaded
			} else {
				// Start with defaults and apply any explicit flags.
				g = genome.Default()
				applyFlags(cmd, g,
					flagName, flagModule, flagRouter, flagDatabase, flagDBTooling,
					flagMigrations, flagObservability, flagLogging, flagConfig,
					flagCI, flagContainer,
				)

				// Only run the TUI wizard if there are missing required fields.
				// When all flags are provided, skip the wizard entirely.
				if g.Project.Name == "" || g.Project.Module == "" {
					if err := cli.RunWizard(g); err != nil {
						return err
					}
				}
			}

			// Validate the genome.
			if err := g.Validate(); err != nil {
				return fmt.Errorf("genome validation failed: %w", err)
			}

			// Load profile.
			reg := profile.NewRegistry()
			if devProfileDir != "" {
				devFS := os.DirFS(devProfileDir)
				p := profile.NewFSProfile("dev", "development profile", devFS)
				reg.Register(p)
			} else {
				templates.RegisterBuiltins(reg)
			}

			profileName := g.Choices.Structure + "-" + g.Choices.Transport
			if devProfileDir != "" {
				profileName = "dev"
			}
			p, err := reg.Get(profileName)
			if err != nil {
				return fmt.Errorf("loading profile: %w", err)
			}

			// Determine output directory.
			if outputDir == "" {
				outputDir = filepath.Join(".", g.Project.Name)
			}
			absOutput, err := filepath.Abs(outputDir)
			if err != nil {
				return fmt.Errorf("resolving output path: %w", err)
			}

			// Check output directory.
			if info, err := os.Stat(absOutput); err == nil && info.IsDir() {
				if !force {
					return fmt.Errorf("output directory %s already exists (use --force to overwrite)", absOutput)
				}
				if err := os.RemoveAll(absOutput); err != nil {
					return fmt.Errorf("removing existing output directory: %w", err)
				}
			}

			// Render via engine.
			eng := engine.New(engine.WithVerbose(verbose))
			if err := eng.Render(g, p, absOutput); err != nil {
				return fmt.Errorf("rendering templates: %w", err)
			}

			// Collect generated files for tree display.
			var generated []string
			_ = filepath.WalkDir(absOutput, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				rel, _ := filepath.Rel(absOutput, path)
				generated = append(generated, rel)
				return nil
			})

			fmt.Println()
			cli.PrintTree(absOutput, generated)
			fmt.Println()

			// Run verification.
			fmt.Println("Running verification...")
			results, err := verify.Run(absOutput)
			if err != nil {
				fmt.Fprintf(os.Stderr, "verification error: %v\n", err)
			} else {
				allPassed := true
				for _, r := range results {
					status := "PASS"
					if !r.Passed {
						status = "FAIL"
						allPassed = false
					}
					fmt.Printf("  [%s] %s\n", status, r.Step)
					if !r.Passed && verbose {
						fmt.Println(r.Output)
					}
				}
				if allPassed {
					fmt.Println("All verification steps passed.")
				}
			}

			// Print next steps.
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  cd %s\n", g.Project.Name)
			fmt.Printf("  go mod tidy\n")
			fmt.Printf("  go run ./cmd/app\n")

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&fromFile, "from", "", "load genome from YAML file (skips TUI)")
	flags.StringVar(&outputDir, "output", "", "output directory (defaults to ./{name})")
	flags.BoolVar(&force, "force", false, "overwrite existing output directory")
	flags.StringVar(&devProfileDir, "dev-profile-dir", "", "load profile from disk instead of embedded")
	flags.BoolVar(&verbose, "verbose", false, "enable verbose output during generation")

	flags.StringVar(&flagName, "name", "", "project name")
	flags.StringVar(&flagModule, "module", "", "Go module path")
	flags.StringVar(&flagRouter, "router", "", "HTTP router (chi, stdlib)")
	flags.StringVar(&flagDatabase, "database", "", "database backend (postgres, sqlite, none)")
	flags.StringVar(&flagDBTooling, "db-tooling", "", "DB tooling (sqlc, none)")
	flags.StringVar(&flagMigrations, "migrations", "", "migration tool (goose, golang-migrate, none)")
	flags.StringVar(&flagObservability, "observability", "", "observability stack (otel-full, otel-traces-only, none)")
	flags.StringVar(&flagLogging, "logging", "", "logging library (slog, zap)")
	flags.StringVar(&flagConfig, "config", "", "config strategy (env, koanf)")
	flags.StringVar(&flagCI, "ci", "", "CI pipeline (github-actions, none)")
	flags.StringVar(&flagContainer, "container", "", "container config (dockerfile, none)")

	return cmd
}

// applyFlags sets genome fields for any flags that were explicitly provided.
func applyFlags(cmd *cobra.Command, g *genome.Genome,
	name, module, router, database, dbTooling, migrations,
	observability, logging, config, ci, container string,
) {
	if cmd.Flags().Changed("name") {
		g.Project.Name = name
	}
	if cmd.Flags().Changed("module") {
		g.Project.Module = module
	}
	if cmd.Flags().Changed("router") {
		g.Choices.Router = router
	}
	if cmd.Flags().Changed("database") {
		g.Choices.Database = database
	}
	if cmd.Flags().Changed("db-tooling") {
		g.Choices.DBTooling = dbTooling
	}
	if cmd.Flags().Changed("migrations") {
		g.Choices.Migrations = migrations
	}
	if cmd.Flags().Changed("observability") {
		g.Choices.Observability = observability
	}
	if cmd.Flags().Changed("logging") {
		g.Choices.Logging = logging
	}
	if cmd.Flags().Changed("config") {
		g.Choices.Config = config
	}
	if cmd.Flags().Changed("ci") {
		g.Choices.CI = ci
	}
	if cmd.Flags().Changed("container") {
		g.Choices.Container = container
	}
}

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [axis]",
		Short: "List available options for each genome axis",
		Long:  "Without arguments, list all axes and their options. With an axis name, list options for that axis.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return listAxis(args[0])
			}
			return listAll()
		},
	}
	return cmd
}

func listAll() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "AXIS\tOPTIONS")
	for _, axis := range genome.AllAxes() {
		values := genome.AllowedValues(axis)
		fmt.Fprintf(w, "%s\t%s\n", axis, strings.Join(values, ", "))
	}
	return w.Flush()
}

func listAxis(axis string) error {
	values := genome.AllowedValues(axis)
	if values == nil {
		return fmt.Errorf("unknown axis: %q\nAvailable axes: %s", axis, strings.Join(genome.AllAxes(), ", "))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Options for %s:\n", axis)
	for _, v := range values {
		implemented := ""
		if genome.IsImplemented(axis, v) {
			implemented = "(implemented)"
		}
		fmt.Fprintf(w, "  %s\t%s\n", v, implemented)
	}
	return w.Flush()
}

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long:  "Generate shell completion scripts for bash, zsh, or fish.",
		Args:  cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}
