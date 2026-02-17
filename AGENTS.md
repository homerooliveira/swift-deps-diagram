# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go CLI for generating Swift dependency diagrams.
- `cmd/swift-deps-diagram/`: CLI entrypoint (`main.go`) and CLI-focused tests.
- `internal/app/`: orchestration pipeline from input resolution to rendering/output.
- `internal/*`: focused modules (`inputresolve`, `swiftpm`, `xcodeproj`, `graph`, `render`, `output`, `graphviz`, `errors`).
- `internal/testutil/`: shared test helpers.
- `testdata/fixtures/`: JSON fixtures used by unit tests.
- `docs/`: architecture and feature docs.

Keep new code in `internal/<domain>` with matching tests, and keep `cmd/` thin.

## Build, Test, and Development Commands
- `go build ./cmd/swift-deps-diagram`: build the CLI binary.
- `go run ./cmd/swift-deps-diagram --format mermaid`: run locally without installing.
- `go test ./...`: run the full test suite.
- `go test ./internal/render -run TestDot`: run a focused test subset.
- `gofmt -w .`: format all Go files before committing.

If you use PNG output (`--format png`), ensure Graphviz `dot` is installed. SwiftPM/Xcode flows also depend on Apple tooling (`swift`, `plutil`).

## Coding Style & Naming Conventions
Follow standard Go conventions:
- Use `gofmt` formatting (tabs for indentation).
- Package names are short, lowercase, no underscores.
- Exported identifiers use `PascalCase`; internal helpers use `camelCase`.
- Keep functions small and single-purpose; return typed errors from `internal/errors` when relevant.
- Name files by feature (`build.go`, `build_test.go`).

## Testing Guidelines
- Place tests beside source files as `*_test.go`.
- Use descriptive test names (`TestDotDeterministicOutput` style).
- Prefer table-driven tests for multiple scenarios.
- Reuse fixtures from `testdata/fixtures/` and helpers from `internal/testutil/`.
- Add/adjust tests with every behavior change; no fixed coverage threshold is enforced.

## Commit & Pull Request Guidelines
Recent commits use concise, imperative subjects (e.g., `Add optional PNG generation via Graphviz`).
- Commit message: one clear action, scoped to a single change.
- PRs should include: purpose, key implementation notes, and test evidence (`go test ./...` output).
- Link related issues when applicable and include sample CLI output when behavior/format changes.
