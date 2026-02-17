# Step 09 - Output Writer and Exit Codes

## Goal
Finalize content emission and consistent process exit behavior.

## Actions
- Add `internal/output/write.go`:
  - write to stdout when `--output` is empty
  - write atomically to file when provided (temp file + rename)
- In app flow:
  - render `mermaid|dot|both`
  - for `both`, join as:
    - `mermaid + "\n\n---\n\n" + dot`
- Add typed errors in `internal/errors/errors.go`.
- Ensure stderr is used for errors, stdout for diagram payloads only.

## Deliverables
- stable IO behavior
- clear error to exit-code mapping

## Unit Tests
- `TestWriteToStdoutWhenNoOutputPath`: empty output path writes content to stdout sink.
- `TestWriteToFileAtomic`: file output uses temp file + rename and produces exact content.
- `TestRenderBothFormatSeparator`: combined format includes `\n\n---\n\n` separator exactly once.
- `TestErrorToExitCodeMapping`: typed errors map consistently to `1` or `2`.
- `TestStdoutOnlyContainsDiagram`: error flows do not pollute stdout payload.

## Done Criteria
- output content is identical between stdout and file modes.
- exit code contract is enforced (`0`, `1`, `2`).
