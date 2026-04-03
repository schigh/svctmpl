//go:build tools

// Package tools pins development tool dependencies for this project.
// These are built into tools/bin/ by `make tools` and are NOT installed globally.
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/pressly/goose/v3/cmd/goose"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)
