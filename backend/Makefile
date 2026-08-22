SHELL := /bin/bash
.DEFAULT_GOAL := help

DATABASE_URL ?= postgres://maxai:maxai@localhost:5432/maxai?sslmode=disable
TEST_DATABASE_URL ?= postgres://maxai:maxai@localhost:5432/maxai_test?sslmode=disable
TEST_REDIS_URL ?= redis://localhost:6379/1
MIGRATIONS_DIR := migrations

.PHONY: help
help: ## Show available targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ----------------------------------------------------------------- local infra

.PHONY: up
up: ## Start PostgreSQL and Redis
	docker compose up -d --wait

.PHONY: down
down: ## Stop containers, keep volumes
	docker compose down

.PHONY: reset
reset: ## Stop containers and delete volumes
	docker compose down -v

# ------------------------------------------------------------------ migrations

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: migrate-status
migrate-status: ## Print the current schema version
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: migrate-create
migrate-create: ## Create a migration pair: make migrate-create name=add_something
	@test -n "$(name)" || (echo "usage: make migrate-create name=add_something" && exit 1)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# ------------------------------------------------------------------------ sqlc

.PHONY: sqlc
sqlc: ## Regenerate database access code from queries/
	sqlc generate

.PHONY: sqlc-vet
sqlc-vet: ## Validate queries against the schema
	sqlc vet

# ------------------------------------------------------------------- processes

.PHONY: run-api
run-api: ## Run the HTTP API
	go run ./cmd/api

.PHONY: run-worker
run-worker: ## Run the background worker
	go run ./cmd/worker

.PHONY: build
build: ## Build both binaries into bin/
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

# ----------------------------------------------------------------------- tests

.PHONY: test
test: ## Run unit tests
	go test ./...

.PHONY: test-integration
test-integration: ## Run integration tests against a live database and Redis
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" TEST_REDIS_URL="$(TEST_REDIS_URL)" \
		go test -tags=integration -count=1 ./tests/...

.PHONY: openapi
openapi: ## Validate the API contract
	go test ./openapi/...

.PHONY: cover
cover: ## Run unit tests and open the coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# ---------------------------------------------------------------------- checks

.PHONY: fmt
fmt: ## Format the module
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Tidy module dependencies
	go mod tidy

.PHONY: check
check: fmt vet test ## Format, vet and unit test
