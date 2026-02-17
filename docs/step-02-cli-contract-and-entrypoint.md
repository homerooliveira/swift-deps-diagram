# Step 02 - CLI Contract and Entrypoint

## Goal
Implement command-line parsing and top-level execution flow.

## Actions
- In `cmd/swift-deps-diagram/main.go`, parse flags:
  - `--path` default `.`
  - `--format` default `both` (`mermaid|dot|both`)
  - `--output` optional
  - `--include-tests` default `false`
- Validate flag values and map failures to exit code `1`.
- Call `internal/app.Run(...)` and map runtime failures to exit code `2`.

## Deliverables
- argument parser
- exit code handling
- usage/help text

## Unit Tests
- `TestParseFlagsDefaults`: no args yields path `.`, format `both`, no output, include-tests false.
- `TestParseFlagsInvalidFormat`: invalid format returns usage error and maps to exit code `1`.
- `TestMainMapsRunErrorToExitCode2`: runtime app error maps to exit code `2`.
- `TestHelpTextIncludesFlags`: `-h` output includes all supported flags.

## Done Criteria
- invalid format exits with code `1`.
- valid invocation calls app runner.
