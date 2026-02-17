# Capabilities Specification

This document defines the behavior of `swift-deps-diagram` as a language-agnostic functional specification.

## 1. Purpose and Scope

`swift-deps-diagram` is a command-line tool that builds a dependency graph and renders it as Mermaid, Graphviz DOT, PNG, or terminal ASCII tree output.

Supported source ecosystems:
- SwiftPM (`Package.swift` via `swift package dump-package`)
- Xcode (`.xcodeproj` and `.xcworkspace`)
- Bazel (`WORKSPACE`, `WORKSPACE.bazel`, or `MODULE.bazel`)

Scope of this specification:
- External CLI contract
- Input resolution behavior
- Canonical graph/data semantics
- Rendering and output semantics
- Error taxonomy and exit code mapping
- Determinism and parity guarantees

## 2. External CLI Contract

### 2.1 Flags, defaults, constraints

| Flag | Type | Default | Meaning |
|---|---|---:|---|
| `--path` | string | `.` | Root path for input detection and source loading |
| `--project` | string | `` | Explicit `.xcodeproj` path |
| `--workspace` | string | `` | Explicit `.xcworkspace` path |
| `--bazel-targets` | string | `` | Bazel query scope expression |
| `--mode` | enum | `auto` | `auto|spm|xcode|bazel` |
| `--format` | enum | `png` | `mermaid|dot|png|terminal` |
| `--output` | string | `` | Output file path (text formats use stdout when empty) |
| `--verbose` | bool | `false` | Print generation details for file outputs |
| `--include-tests` | bool | `false` | Include test targets/rules in the graph |

Constraints:
- `--project` and `--workspace` are mutually exclusive.
- Positional arguments are rejected.
- Invalid `--mode` or `--format` values are rejected.
- In `spm` mode, provided Xcode-path flags are ignored with a warning.

### 2.2 Mode/format validation rules

Validation order:
1. Parse flags.
2. Validate `format ∈ {mermaid,dot,png,terminal}`.
3. Validate `mode ∈ {auto,spm,xcode,bazel}`.
4. Validate that `--project` and `--workspace` are not both set.
5. Validate there are no positional arguments.

### 2.3 Exit code contract

| Exit code | Meaning |
|---:|---|
| `0` | Success |
| `1` | Invalid input/arguments or missing project markers |
| `2` | Runtime/tool/parse/render/output failure |

## 3. Input Resolution Rules (Normative)

### 3.1 Mode-specific resolver behavior

| Requested mode | Resolver behavior |
|---|---|
| `spm` | Requires `Package.swift`; returns SPM resolution with package path |
| `xcode` | Resolves project/workspace; returns Xcode resolution |
| `bazel` | Requires Bazel workspace marker; returns Bazel workspace + normalized target scope |
| `auto` | Applies precedence: Xcode, then Bazel, then SwiftPM |

### 3.2 `auto` precedence and tie-breakers

Precedence in `auto` mode:
1. Xcode
2. Bazel
3. SwiftPM

Tie-breakers and details:
- If Xcode flags are explicitly provided in `auto`, resolver forces Xcode resolution.
- For directory scanning:
  - Choose first lexicographically sorted `.xcworkspace` if any.
  - Otherwise choose first lexicographically sorted `.xcodeproj` if any.
- If Xcode does not resolve, check Bazel markers.
- If Bazel does not resolve, check `Package.swift`.

### 3.3 Explicit `--project` / `--workspace` behavior

`--project`:
- Must resolve to an existing `.xcodeproj` path.
- Returns project path directly.

`--workspace`:
- Must resolve to an existing `.xcworkspace` path.
- Attempts to find project from `contents.xcworkspacedata` references first.
- Supported reference prefixes: `group:`, `container:`, `self:`, `absolute:`.
- Fallback: first lexicographically sorted `.xcodeproj` in workspace parent directory.

### 3.4 Bazel marker detection

Workspace markers:
- `WORKSPACE`
- `WORKSPACE.bazel`
- `MODULE.bazel`

Rules:
- If input is one of these files, workspace is its parent directory.
- If input is a directory, marker presence in that directory qualifies as a Bazel workspace.
- Empty Bazel target scope is normalized to `//...`.

### 3.5 Errors for “nothing found”

If `auto` cannot resolve Xcode, Bazel, or SwiftPM markers, resolver returns input-not-found with a combined message indicating all checked marker classes.

## 4. Canonical Data Model

### 4.1 Graph model

Canonical graph structure:
- `Graph { Nodes, Edges }`
- `Node { ID, Label, Kind }`
- `Edge { FromID, ToID, Kind }`

Kinds:
- Node kinds: `target`, `external_product`
- Edge kinds: `target`, `product`, `by_name`

### 4.2 ID schema (normative)

| Entity class | ID schema | Notes |
|---|---|---|
| Swift target | `target::<name>` | Xcode duplicate-name targets may be suffixed with target identifier |
| Package product | `pkg::<package>::<product>` | Used when package identity is known |
| Product without package identity | `product::<name>` | Used when package identity is unknown |
| byName unresolved symbol | `name::<name>` | byName fallback when local target does not exist |
| Bazel local target | `target::<label>` | Label includes `//...` |
| Bazel external dep | `external::<label>` | Label form `@repo//...` |

### 4.3 Deterministic ordering guarantees

Deterministic behavior:
- Node traversal order is lexicographic by node ID.
- Edge traversal order is lexicographic by `(FromID, ToID, Kind)`.
- Builders return deduplicated, sorted edges.
- Renderers consume sorted nodes/edges.

## 5. Source Adapters

### 5.1 SwiftPM adapter behavior

Behavior:
- Requires `swift` available in `PATH`.
- Executes `swift package dump-package --package-path <path>`.
- Uses 30-second timeout.
- Includes command stderr details in failure messages.
- Empty stdout is treated as failure.

Failure classes:
- Swift tool missing.
- Dump command timeout/failure/empty output.

### 5.2 Xcode adapter behavior

Behavior:
- Requires `plutil` available in `PATH`.
- Validates project path and pbxproj presence.
- Converts pbxproj to JSON using `plutil`.
- Extracts:
  - Build targets
  - Target dependencies
  - Proxy-based dependencies
  - Swift package product dependencies
  - Package identity from explicit identity or repository URL/path derivation

Failure classes:
- Project/workspace not found or structurally invalid.
- pbxproj parse/conversion failure.

### 5.3 Bazel adapter behavior

Behavior:
- Resolves Bazel executable by trying `bazel`, then `bazelisk`.
- Normalizes empty scope to `//...`.
- Query steps:
  1. Rule labels in scope.
  2. Rule kinds for labels.
  3. Direct deps for each target using `deps(label, 1)`.
- Adds `--noimplicit_deps` and `--notool_deps`.
- Filters deps to local (`//...`) and external (`@...`) labels.
- Removes self-dependencies.
- Deduplicates and sorts labels/deps.

Failure classes:
- Bazel binary not found.
- Query timeout/failure.
- Rule-kind parse failure.

## 6. Graph Construction Semantics

### 6.1 SwiftPM target/product/byName resolution

Rules:
- Include each non-filtered target as a target node.
- Dependency mapping:
  - `target`: link to local target if present; otherwise fallback to external product node.
  - `product`: always external product node.
  - `byName`: local target if present; otherwise `name::<name>` external-style node.
- Empty dependency names are ignored.
- Unknown dependency kinds are ignored.

### 6.2 Test filtering behavior (`--include-tests`)

When `--include-tests` is false:
- SwiftPM excludes targets with type `test`.
- Xcode excludes targets whose product type indicates unit/UI test bundles.
- Bazel excludes rule kinds ending with `_test`.

When true, these are included.

### 6.3 Edge deduplication

All graph builders deduplicate edges by `(Kind, FromID, ToID)`.

### 6.4 External dependency handling

- SwiftPM/Xcode external package products are represented as `external_product` nodes.
- Bazel `@repo` dependencies are represented as external nodes (`external::<label>`).

## 7. Rendering Semantics

### 7.1 Mermaid output contract

Output rules:
- Header: `flowchart TD`
- Deterministic synthetic node IDs: `n1`, `n2`, ... by sorted canonical node order.
- Node line shape: `nX["label"]`
- Edge line shape: `nX --> nY`

Label escaping:
- Remove backticks.
- Replace newlines with spaces.
- Escape double quotes.

### 7.2 DOT output contract and styling

Output rules:
- Header: `digraph dependencies {`
- Orientation: `rankdir=TB` (top-to-bottom).
- Target node style: `shape=box`.
- External node style: `shape=ellipse,style=dashed`.
- Directed edges rendered with `->`.

Escaping:
- Escape backslashes and double quotes.
- Replace newlines with spaces.
- Always quote IDs and labels.

### 7.3 PNG generation flow

PNG flow:
1. Build canonical graph.
2. Render DOT text.
3. Determine output path (`--output` or default `deps.png`).
4. Invoke Graphviz `dot -Tpng -o <path>` with DOT on stdin.

PNG adapter behavior:
- Empty output path is a no-op.
- Missing `dot` is a specific failure class.
- Rendering failures include stderr details.
- Render timeout is 30 seconds.

### 7.4 Terminal output contract

Output rules:
- Render roots as ASCII tree blocks.
- Roots are target nodes with no incoming dependency from another target.
- If no zero-incoming roots exist, choose deterministic fallback roots to cover unresolved target dependency components.
- Child ordering is deterministic by child label, child ID, and edge kind.
- Shared nodes in different branches are rendered in each branch.
- Cyclic back-references are shown as `(*)` and not expanded again.
- Empty render result is exactly `(empty)`.

## 8. Output and Logging Behavior

### 8.1 stdout vs file output

Text formats (`mermaid`, `dot`, `terminal`):
- If output path is empty, write text to stdout.
- Otherwise write to file.

PNG format:
- Produces a PNG file; default path is `deps.png` when output path is empty.
- Does not emit graph text to stdout.

### 8.2 Atomic file write semantics

For text file output:
1. Ensure destination directory exists.
2. Create temporary file in destination directory.
3. Write content.
4. Close file.
5. Rename temp file to final path.
6. Best-effort temp cleanup.

### 8.3 PNG success message behavior

Committed behavior contract:
- PNG success message is emitted on success.
- Message format: `generated png using dot format at <path>`.

Also:
- Verbose file-output messages exist for `mermaid`, `dot`, and `terminal` file outputs.
- No verbose message for `mermaid`/`dot`/`terminal` when writing to stdout.

## 9. Error Taxonomy

### 9.1 Error kinds and origin categories

| Error kind | Origin category |
|---|---|
| `invalid_args` | CLI parsing/validation and runtime option validation |
| `manifest_not_found` | SwiftPM marker/path resolution |
| `input_not_found` | Auto-mode marker resolution |
| `ambiguous_input` | Reserved input classification |
| `swift_not_found` | SwiftPM tool discovery |
| `dump_package_failed` | SwiftPM command execution |
| `manifest_decode_failed` | Manifest decoding |
| `xcode_project_not_found` | Xcode path/project discovery |
| `xcode_parse_failed` | Xcode pbxproj conversion/parsing |
| `xcode_unsupported_structure` | Reserved Xcode classification |
| `bazel_workspace_not_found` | Bazel marker/workspace discovery |
| `bazel_binary_not_found` | Bazel tool discovery |
| `bazel_query_failed` | Bazel query execution/timeout |
| `bazel_parse_failed` | Bazel query output parsing |
| `graphviz_not_found` | Graphviz tool discovery |
| `graphviz_render_failed` | Graphviz render failure/timeout |
| `output_write_failed` | File/stdout write failures |
| `runtime_failed` | Generic orchestration failure wrapper |

### 9.2 Error-kind to exit-code mapping

| Error kind group | Exit code |
|---|---:|
| Invalid-args/input-location class | `1` |
| Runtime/tool/parse/render/output class | `2` |

## 10. Determinism and Portability Guarantees

Must remain stable in a parity port:
- Canonical node and edge ordering.
- Edge deduplication semantics.
- Resolver precedence and tie-breakers.
- Renderer escaping rules and output structure.
- Default output path behavior for PNG.
- Exit code mapping classes.

Allowed to vary by environment:
- External tool availability and underlying tool error strings.
- Filesystem/platform-specific low-level error details.

## 11. Non-Goals / Known Limits

Current behavior intentionally does not provide:
- Full semantic validation of every field in source project formats.
- Full transitive Bazel closure (uses direct deps depth=1 per target query).
- Cycle analysis or advanced graph analytics.
- Built-in PNG renderer independent of Graphviz.
- Recovery from arbitrary malformed external-tool output beyond typed failure signaling.
