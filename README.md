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

## License

Apache 2.0. See [LICENSE](LICENSE) for details.
