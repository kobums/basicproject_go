# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go REST API backend boilerplate (basicproject_go). Designed to act as a starter template for new backend projects. Built with Fiber v2 web framework on Go 1.26.

## Build & Run Commands

```bash
make server        # Build binary (runs go build)
make test          # Run tests: go test -v ./...
make run           # Run via go run main.go
make linux         # Cross-compile for Linux
make dockerbuild   # Build static Linux binary for Docker
make docker        # Build Docker image (kobums/basicproject_go)
make clean         # Remove built binary
```

The build process expects `buildtool-model` and `buildtool-router` to auto-generate model and router code from `model.json`.

## Architecture

**MVC-style layered architecture:**

- **`main.go`** — Entry point
- **`services/http.go`** — Fiber app setup (CORS, compression, logging, static files, routes)
- **`router/`** — Route definitions
- **`controllers/`** — `controllers.go` defines base `Controller`. Domain controllers generated in `rest/` embed this base.
- **`models/`** — Database setup (`db.go`)
- **`config/`** — Parses `model.json`
- **`global/`** — Shared utilities

## Adding a New Domain

1. Define table and structure in `model.json`
2. Run code generation tools
3. Add table to `basicproject_go.sql`
4. Re-run `make server` or `make run`
