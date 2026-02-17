# Step 04 - Manifest JSON Models and Decoder

## Goal
Decode only the required subset of `dump-package` JSON safely and predictably.

## Actions
- Add `internal/manifest/model.go` types for:
  - package name
  - targets
  - target type
  - dependency variants (`target`, `product`, `byName`)
- Add `internal/manifest/decode.go` with:
  - `Decode(data []byte) (Package, error)`
- Keep decoder forward-compatible (do not reject unknown fields).
- Normalize dependency shapes into a single internal representation.

## Deliverables
- manifest model types
- decode entrypoint returning typed package struct

## Unit Tests
- `TestDecodeManifestMinimal`: minimal valid fixture decodes package and targets.
- `TestDecodeManifestDependenciesVariants`: fixture with `target`, `product`, `byName` decodes correctly.
- `TestDecodeManifestUnknownFieldsIgnored`: extra JSON fields do not break decode.
- `TestDecodeManifestInvalidJSON`: malformed JSON returns decode error.

## Done Criteria
- fixture JSON decodes successfully.
- malformed JSON returns a decode error with context.
