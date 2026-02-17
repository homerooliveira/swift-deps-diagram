# Step 10 - Tests, Fixtures, and Acceptance

## Goal
Lock correctness with unit and integration coverage and verify acceptance criteria.

## Actions
- Add fixtures in `testdata/fixtures`:
  - simple local target graph
  - target + product dependencies
  - byName local and external fallback
  - optional test target cases
- Unit tests:
  - manifest decode
  - graph builder rules and dedup
  - Mermaid renderer snapshots
  - DOT renderer snapshots
  - output writer behavior
- Integration tests for app runner:
  - valid run per format
  - invalid args -> code `1`
  - dump failure/decode failure -> code `2`
- Document usage and troubleshooting in `README.md`.

## Deliverables
- test suite and fixtures
- final CLI usage documentation

## Unit Tests
- `TestFixtureSimpleLocalGraph`: validates decode/build/render baseline fixture.
- `TestFixtureTargetPlusProduct`: validates external product node rendering in both formats.
- `TestFixtureByNameLocalAndExternal`: validates byName resolution path and fallback.
- `TestAppRunnerMermaidMode`: integration-style unit test for `--format mermaid`.
- `TestAppRunnerDotMode`: integration-style unit test for `--format dot`.
- `TestAppRunnerBothMode`: integration-style unit test for `--format both`.
- `TestAppRunnerInvalidArgsExit1`: invalid flags map to exit code `1`.
- `TestAppRunnerRuntimeFailureExit2`: dump/decode/output failures map to exit code `2`.

## Done Criteria
- `go test ./...` passes.
- manual smoke run against a real Swift package emits expected diagrams.
- acceptance checklist below is all true.

## Acceptance Checklist
- CLI reads package manifest via `swift package dump-package`.
- Target-level dependencies are represented correctly.
- Mermaid and DOT outputs are both supported.
- Default output is stdout; `--output` writes files.
- Output ordering is deterministic.
- Error messages are actionable and exit codes are stable.
