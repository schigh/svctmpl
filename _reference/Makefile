BINARY   := myservice
PKG      := ./cmd/app
GOFLAGS  := -ldflags="-s -w"
TOOLS_BIN := tools/bin

.PHONY: build run test lint migrate-up migrate-down generate \
        docker-build docker-run clean tools help \
        check-golangci-lint check-goose check-sqlc check-docker

## build: compile the service binary
build:
	go build $(GOFLAGS) -o $(BINARY) $(PKG)

## run: build and run the service
run: build
	./$(BINARY)

## test: run tests with race detector
test:
	go test -race -v -count=1 ./...

## lint: run golangci-lint (project-local binary)
lint: $(TOOLS_BIN)/golangci-lint
	$(TOOLS_BIN)/golangci-lint run ./...

## migrate-up: apply database migrations
migrate-up: $(TOOLS_BIN)/goose
	$(TOOLS_BIN)/goose -dir migrations postgres "$$DATABASE_URL" up

## migrate-down: roll back the last migration
migrate-down: $(TOOLS_BIN)/goose
	$(TOOLS_BIN)/goose -dir migrations postgres "$$DATABASE_URL" down

## generate: regenerate sqlc query code
generate: $(TOOLS_BIN)/sqlc
	$(TOOLS_BIN)/sqlc generate

## docker-build: build the Docker image
docker-build: check-docker
	docker build -t $(BINARY):latest .

## docker-run: build and run the Docker image
docker-run: docker-build
	docker run --rm -p 8080:8080 -p 9090:9090 --env-file .env $(BINARY):latest

## clean: remove build artifacts and test cache
clean:
	rm -f $(BINARY)
	go clean -testcache

## tools: build all development tools into tools/bin/
tools: $(TOOLS_BIN)/golangci-lint $(TOOLS_BIN)/goose $(TOOLS_BIN)/sqlc
	@echo "Tools built to $(TOOLS_BIN)/"

$(TOOLS_BIN)/golangci-lint: tools/go.sum
	cd tools && go build -o ../$(TOOLS_BIN)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(TOOLS_BIN)/goose: tools/go.sum
	cd tools && go build -o ../$(TOOLS_BIN)/goose github.com/pressly/goose/v3/cmd/goose

$(TOOLS_BIN)/sqlc: tools/go.sum
	cd tools && go build -o ../$(TOOLS_BIN)/sqlc github.com/sqlc-dev/sqlc/cmd/sqlc

tools/go.sum: tools/go.mod tools/tools.go
	cd tools && go mod tidy

## check-docker: verify docker is available
check-docker:
	@command -v docker >/dev/null 2>&1 || { echo "docker not found. Install from https://docs.docker.com/get-docker/"; exit 1; }

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
