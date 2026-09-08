# Repository Instructions

## Code Boundaries

- Use Go 1.27 or later, as declared in `go.mod`, `tests/go.mod`, and `go.work`. The workspace contains two modules: the root module and `tests/`. Root `go test ./...` does not include the E2E module.
- Most plugin implementations are external Go modules. `container/plugins.go:Plugins()` defines the default plugin set for the Endure container. Add bundled plugins there.
- The CLI starts at `cmd/rr/main.go` and `internal/cli/root.go`. Lifecycle changes can affect three separate implementations: `internal/cli/serve/command.go` (non-Windows), `internal/cli/serve/command_windows.go` (Windows), and `lib/roadrunner.go` (embedding API).

## Build And Checks

- From the root, `make build` creates `./rr`. It sets `CGO_ENABLED=0` for the build only. Keep CGO enabled for race tests.
- Root tests: `make test` runs `go test -v -race ./...`. Focus on one package with `go test -v -race ./internal/rpc`. Select one test with `go test -v -race ./internal/rpc -run '^TestNewClient_WithIncludes$'`.
- Root lint: `golangci-lint run -v --build-tags=race --timeout=10m`, matching `.github/workflows/tests.yml`. Use golangci-lint v2 with `.golangci.yml`.
- E2E setup matches `.github/workflows/e2e.yml`: Ubuntu, PHP 8.5 with the `sockets` extension, and Composer. Run `composer update --prefer-dist --no-progress --ansi` from `tests/php_test_files/`. The root `composer.json` is only a metapackage; it does not install the worker dependencies.
- From `tests/`, run `go mod download`, then `go test -timeout 15m -v -race -failfast ./...`. For one E2E test, use `go test -timeout 15m -v -race -run '^TestHTTPWithMiddleware$' .`. Fixtures are in `tests/configs/` and `tests/php_test_files/`.
- RPC and E2E tests use fixed local TCP ports. Library tests share `os.TempDir() + "/.rr.yaml"`. Avoid simultaneous runs of the same tests.

## Runtime Configuration

- Root `.rr.yaml` is an options reference with placeholder worker commands, not a ready-to-run application config. Config files use `version: '3'`.
- The CLI changes to the config file's directory unless `-w` is set. With `-w`, it resolves `-c` from that working directory. It loads dotenv after this directory change. `DOTENV_PATH` takes precedence over `--dotenv` (`internal/cli/root.go`).
- `make debug` refers to the missing `.rr-sample-bench-http.yaml`. Use `dlv debug cmd/rr/main.go -- serve -c <config>` with an existing configuration.

## Public Schemas

- Do not rename or remove `schemas/` or any path inside it. These are public schema URLs; see `schemas/readme.md`.
- From `schemas/`, run `npm install`, then `node test.js` to validate root `.rr.yaml` against `config/3.0.schema.json`. `npm test` is a placeholder that exits with failure. The validator resolves remote plugin schema references and needs network access.

## Contribution Requirements

- `.github/pull_request_template.md` requires commit sign-off (`git commit -s`) and `CHANGELOG.md` entries for user-facing changes.
