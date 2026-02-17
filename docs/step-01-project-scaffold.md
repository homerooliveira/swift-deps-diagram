# Step 01 - Project Scaffold

## Goal
Create a Go CLI repository skeleton ready for implementation and testing.

## Actions
- Initialize module: `go mod init swift-deps-diagram`.
- Create directories:
  - `cmd/swift-deps-diagram`
  - `internal/app`
  - `internal/swiftpm`
  - `internal/manifest`
  - `internal/graph`
  - `internal/render`
  - `internal/output`
  - `internal/errors`
  - `testdata/fixtures`
- Add a placeholder `main.go` and package stubs.

## Deliverables
- `go.mod`
- empty package files per directory
- compilable placeholder binary

## Unit Tests
- `TestModuleBuilds`: verify `go list ./...` succeeds for all packages.
- `TestMainEntrypointCompiles`: verify `go test ./cmd/swift-deps-diagram` compiles without missing symbols.
- `TestPackageLayoutExists`: verify required directories and placeholder package files exist.

## Done Criteria
- `go test ./...` runs with no package-not-found errors.
- `go build ./cmd/swift-deps-diagram` succeeds.
