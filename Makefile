TEMPL := $(shell go env GOPATH)/bin/templ

.PHONY: generate test run up down build tools

# Regenerate Go from .templ templates. Run after editing any *.templ file.
generate:
	$(TEMPL) generate

# Fast unit tests (in-memory adapters, no container).
test: generate
	go test ./...

# Run locally on the in-memory store — no database needed.
run: generate
	go run ./cmd/web

# Full stack in docker-compose (app + Postgres/pgvector).
up:
	docker compose up --build

down:
	docker compose down

build: generate
	go build ./...

# Install build-time tooling (templ).
tools:
	go install github.com/a-h/templ/cmd/templ@latest
