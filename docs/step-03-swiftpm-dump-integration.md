# Step 03 - SwiftPM Manifest Dump Integration

## Goal
Load manifest data by calling SwiftPM instead of parsing `Package.swift` text directly.

## Actions
- Add `internal/swiftpm/dump.go` with:
  - `DumpPackage(ctx, packagePath) ([]byte, error)`
- Execute:
  - `swift package dump-package --package-path <path>`
- Capture stdout and stderr separately.
- Add a 30s timeout via `context.WithTimeout`.
- Map missing `swift` binary and command failures to typed errors.

## Deliverables
- reusable SwiftPM command wrapper
- actionable error messages

## Unit Tests
- `TestDumpPackageSuccess`: mocked command runner returns stdout JSON payload.
- `TestDumpPackageSwiftNotFound`: missing binary maps to typed not-found error.
- `TestDumpPackageCommandFailureIncludesStderr`: non-zero exit returns typed failure with stderr snippet.
- `TestDumpPackageRespectsTimeout`: context timeout cancels execution and returns timeout error.

## Done Criteria
- returns JSON bytes for valid package path.
- returns typed error on command failure with stderr context.
