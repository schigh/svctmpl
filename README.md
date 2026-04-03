# svctmpl

Scaffold production-ready Go services from a decision genome.

Instead of copying boilerplate or wiring up the same infrastructure for every new service, `svctmpl` captures your architectural decisions in a single YAML file (a "genome") and generates a complete, compilable, production-shaped Go project from it.

## What it does

You declare choices -- HTTP router, database, observability stack, logging, CI, container strategy -- and `svctmpl` renders a fully wired service with clean architecture, tests, deployment manifests, and a working `docker-compose` stack. Every generated project compiles and passes `go vet` out of the box.

## Install

**Go install** (requires Go 1.25+):

```sh
go install github.com/schigh/svctmpl/cmd/svctmpl@latest
```

**Homebrew** (after the first release):

```sh
brew install schigh/tap/svctmpl
```

**GitHub Releases:**

Download the binary for your platform from the [releases page](https://github.com/schigh/svctmpl/releases).

## Quick start

### Interactive mode

```sh
svctmpl new
```

This launches a TUI wizard that walks you through each decision axis and generates the project.

### From a genome file

```sh
svctmpl new --from genome.yaml
```

Example `genome.yaml`:

```yaml
version: "1"
project:
  name: ordersvc
  module: github.com/acme/ordersvc
choices:
  transport: http
  router: chi
  database: postgres
  db_tooling: sqlc
  migrations: goose
  structure: layered
  observability: otel-full
  logging: slog
  config: env
  ci: github-actions
  container: dockerfile
  compose: true
  k8s: true
  tilt: true
```

### Flags-only mode

Skip the TUI entirely by passing all required values as flags:

```sh
svctmpl new \
  --name ordersvc \
  --module github.com/acme/ordersvc \
  --router chi \
  --database postgres \
  --db-tooling sqlc \
  --migrations goose \
  --observability otel-full \
  --logging slog \
  --config env \
  --ci github-actions \
  --container dockerfile \
  --compose --k8s --tilt
```

## Generated output

A fully scaffolded project looks like this:

```
ordersvc/
  cmd/app/main.go
  internal/
    app/
      app.go
      database.go
      otel.go
      server.go
    config/config.go
    convert/resource.go
    errs/errors.go
    log/log.go
    model/resource.go
    repository/
      db.go
      models.go
      queries.sql
      queries.sql.go
      resource.go
    service/
      repository.go
      resource.go
    transport/http/
      errors.go
      handler.go
      handler_test.go
      health.go
      middleware.go
      resource.go
      resource_test.go
  migrations/001_create_resources.sql
  deploy/
    grafana/datasources.yml
    otel-collector-config.yaml
    prometheus.yml
  k8s/
    configmap.yaml
    deployment.yaml
    hpa.yaml
    ingress.yaml
    kustomization.yaml
    namespace.yaml
    secret.yaml
    service.yaml
  .github/workflows/ci.yml
  .env.example
  .gitignore
  docker-compose.yml
  Dockerfile
  go.mod
  Makefile
  README.md
  sqlc.yaml
  Tiltfile
  tools/
    go.mod
    tools.go
```

## Features

- **Layered architecture** with clean separation: transport, service, repository, model
- **chi router** with middleware (logging, request ID, CORS, recovery)
- **PostgreSQL** with sqlc for type-safe queries and goose for migrations
- **OpenTelemetry** traces, metrics, and logs wired end-to-end with OTel Collector, Prometheus, and Grafana
- **Structured logging** via slog with trace correlation
- **Docker Compose** stack with Postgres, OTel Collector, Prometheus, Jaeger, and Grafana
- **Kubernetes manifests** with Deployment, Service, HPA, Ingress, ConfigMap, and Secrets (Kustomize-ready)
- **Tiltfile** for live-reload development with hot rebuild
- **GitHub Actions CI** with build, vet, test, and lint steps
- **Multi-stage Dockerfile** with distroless final image
- **Health checks** (`/healthz`, `/readyz`) baked in
- **Verification**: generated projects are automatically compiled and vetted after scaffolding

## Available axes

```sh
svctmpl list
```

```
AXIS          OPTIONS
transport     http
router        chi, stdlib
database      postgres, sqlite, none
db_tooling    sqlc, none
migrations    goose, golang-migrate, none
structure     layered
observability otel-full, otel-traces-only, none
logging       slog, zap
config        env, koanf
ci            github-actions, none
container     dockerfile, none
```

## Philosophy

svctmpl is structurally opinionated. The stack choices (which router, which database) are configurable. The architecture is not.

Every generated service follows the same principles:

- **Hexagonal architecture.** Transport delegates to service, service delegates to repository. Each layer defines the interface for the layer below it (consumer-defined interfaces). No circular imports, no leaky abstractions.
- **Transport owns the boundary.** String parsing, request validation, and type conversion happen at the transport layer. The domain model uses real Go types (`uuid.UUID`, not `string`). The service and repository never see raw user input.
- **Modular bootstrapping.** One concern per file in `internal/app/`. Adding Redis means adding `redis.go`, not touching 5 places. Shutdown happens in reverse initialization order.
- **Contextual logging.** Middleware stashes fields (request ID, trace ID) into context. Downstream code calls `log.Ctx(ctx)` and gets a logger with all stashed fields automatically.
- **Sentinel errors.** Cross-cutting error types live in `internal/errs/`. The transport layer maps them to HTTP status codes. The service wraps them with context.
- **Explicit type conversions.** `internal/convert/` holds all type mapping between representations (sqlc models to domain models, etc.). When you see `convert.RepositoryResourceToModel(...)`, you know exactly what's happening.

The goal: code should be boring in structure so it can be interesting in logic. When you need to add a feature, you shouldn't wonder where it goes.

## Roadmap

svctmpl is in active development. Here's where it's going.

### v0.1.0 (current)
- [x] Interactive TUI wizard with Bubbletea/Huh
- [x] Genome YAML schema with validation
- [x] One template profile: layered-http (chi + postgres/sqlc + OTel + slog)
- [x] Hexagonal architecture with consumer-defined interfaces
- [x] Modular app bootstrapping with ordered shutdown
- [x] Context-scoped logging (Stash/Ctx pattern)
- [x] Docker Compose with full OTel stack (Collector + Prometheus + Grafana)
- [x] Kubernetes manifests (Deployment, Service, ConfigMap, Secret, Ingress, HPA)
- [x] Tiltfile for live-reload development
- [x] Post-generation verification (go build + go vet + go test)
- [x] goreleaser for cross-platform distribution

### v0.2.0 — More stack choices
- [ ] [Ent](https://entgo.io) entity framework with codegen (Go-first alternative to sqlc)
- [ ] [Atlas](https://atlasgo.io) declarative schema migrations (pairs with Ent)
- [ ] stdlib router (net/http mux, Go 1.22+)
- [ ] Zap logging option
- [ ] koanf config option
- [ ] SQLite database option
- [ ] golang-migrate migration option
- [ ] Structured health checks via [github.com/schigh/health](https://github.com/schigh/health)

### v0.3.0 — Structure as configuration
- [ ] `svctmpl inspect` — infer a genome from an existing Go project
- [ ] Concern placement as config (where does validation go? where do converters live?)
- [ ] Multiple structure profiles (flat for tiny services, DDD for complex domains)
- [ ] User-supplied template packs loaded from disk

### v1.0.0 — The agentic overseer
- [ ] AST-based code generation (always-valid Go output)
- [ ] `svctmpl evolve` — read genome, diff codebase, apply surgical migrations
- [ ] `svctmpl add <feature>` — add gRPC transport, Redis cache, etc. to existing projects
- [ ] `svctmpl doctor` — verify project health against genome expectations
- [ ] AI agent that understands the genome and generates code with 100% structural certainty

### Contributing

Interested in contributing? The best way right now is to generate a service, use it, and [open an issue](https://github.com/schigh/svctmpl/issues) with what felt wrong or missing. The template structure and genome schema are the areas where feedback has the most impact.

## License

Apache 2.0. See [LICENSE](LICENSE) for details.
