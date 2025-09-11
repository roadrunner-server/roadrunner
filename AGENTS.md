# Repository Guidelines

This guide helps contributors work efficiently on the RoadRunner core (Go) CLI and runtime.

## Project Structure & Module Organization
- `cmd/rr/`: CLI entrypoint (`main.go`) and basic CLI tests.
- `internal/`: CLI commands, debug helpers, metadata, RPC, service wiring.
- `lib/`: Public Go API to embed and control RoadRunner (`RR` type).
- `schemas/`: YAML schemas and config examples; `.rr.yaml` at repo root.
- `benchmarks/`, `container/`: Performance samples and container settings.
- Tests live alongside code as `*_test.go` files.

## Build, Test, and Development Commands
- `make build` — build the `rr` binary to `./rr`.
- `make test` — run `go test -v -race ./...` across modules.
- `./rr serve -c .rr.yaml` — run locally with the sample config.
- `dlv debug cmd/rr/main.go -- serve -c .rr-sample-bench-http.yaml` — debug run (needs Delve).
- `golangci-lint run` — lint/format per `.golangci.yml` (install locally).

## Coding Style & Naming Conventions
- Go 1.x standards: `gofmt`/`goimports`; tabs; 120‑char lines (see linter config).
- Package names: short, lower‑case; exported identifiers use Go’s `UpperCamelCase`.
- Errors: wrap with `%w`; prefer sentinel/typed errors; no panics in library code.
- Keep functions small; avoid globals (see `gochecknoglobals`); prefer context‑aware APIs.

## Testing Guidelines
- Use table‑driven tests; place in `*_test.go`. Call `t.Parallel()` where safe.
- Run with race detector and coverage: `go test -race -cover ./...`.
- Add tests for new CLI flags, config parsing, and plugin wiring. Keep fixtures minimal.

## Commit & Pull Request Guidelines
- Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`, `ci:`.
- PRs must include: clear description, linked issues, test updates, and config/schema changes if applicable.
- Ensure `make test` and `golangci-lint run` pass; include usage examples for CLI‑related changes.

## Security & Configuration Tips
- Never commit secrets; prefer `.env` loaded via `DOTENV_PATH` or `--dotenv`.
- Debug server (`-d`) listens on `:6061`; avoid exposing in production.
