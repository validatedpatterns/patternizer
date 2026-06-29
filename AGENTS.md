# Patternizer - Agent Guidelines

Patternizer is a Go CLI tool that bootstraps Git repositories containing Helm charts into ready-to-use Validated Patterns. Module path: `github.com/validatedpatterns/patternizer`.

## Build, Lint, and Test

Run `make dev-setup` once to install tooling (golangci-lint, ginkgo).

Key make targets:

- `make ci` - Full CI pipeline: lint, build, test. Run this before considering any change complete.
- `make build` - Build the binary.
- `make test` - Run all tests via Ginkgo.
- `make lint` - Run all linters (gofmt, go vet, golangci-lint).
- `make fmt` - Auto-format Go code.
- `make check` - Quick check: format, vet, build, unit tests.

All Go source lives under `src/`. The Makefile handles `cd src` automatically.

## Code Style

- No unnecessary inline comments. Code should be self-documenting through clear naming and structure.
- No emojis, ASCII art, or non-standard characters anywhere in the codebase.
- Exported functions must have godoc comments (enforced by the revive linter).
- Run `make fmt` before committing. Formatting is checked by `make lint`.
- Follow the existing project layout:
  - `src/cmd/` - Cobra CLI command definitions.
  - `src/internal/` - Private packages (not importable externally).
  - `src/internal/types/` - Shared data structures.
  - `src/internal/embedded/` - Embedded resources via `go:embed`.

## Go Best Practices

- Wrap errors with context: `fmt.Errorf("description: %w", err)`.
- Break large functions into smaller, well-named helper functions. Each function should do one thing.
- Keep `main.go` minimal; delegate to `cmd.Execute()`.
- Use `internal/` to keep implementation details private.
- Use `go:embed` for bundled resources (see `internal/embedded/`).
- CLI commands use Cobra. Add new commands by registering them on `rootCmd` in `cmd/root.go`.

## Testing

- Tests must be added or updated when writing new code.
- The project uses Ginkgo v2 and Gomega (BDD style: `Describe`/`Context`/`It`).
- Test files live alongside source files as `*_test.go`.
- Each package with tests needs a `*_suite_test.go` file for Ginkgo bootstrap.
- Use `GinkgoT().TempDir()` for temporary directories in tests.
- Run `make test` to execute all tests (`ginkgo -v ./...`).

## Versioning

The project version is tracked in the shield badge on line 1 of `README.md` (e.g. `![Version: 2.0.0](...)`). Bump it once per session or PR when the CLI binary changes:

- **Patch** (2.0.0 -> 2.0.1): Bug fixes, internal refactors, dependency updates.
- **Minor** (2.0.0 -> 2.1.0): New features, new commands, new flags, or changed behavior.
- **Major**: Ask the user for confirmation before bumping the major version.
- **No bump needed**: Documentation-only changes, CI config, or other changes that do not affect the compiled binary.

Only bump the version once even if multiple code changes are made in the same session.

## Documentation

- Update `README.md` when code changes affect user-facing behavior, CLI usage, or flags.
- Keep the README focused on usage and contribution guidance.

## Verification

Run `make ci` before submitting any change. It runs the full pipeline:

1. Lint (gofmt, go vet, golangci-lint with gocritic, misspell, and revive)
2. Build
3. Test (all Ginkgo suites)
