# swift-deps-diagram

CLI tool to generate dependency diagrams from a Swift Package manifest (`Package.swift`) by using `swift package dump-package`.
Also supports Xcode projects/workspaces and Bazel workspaces.

## Build

```bash
go build ./cmd/swift-deps-diagram
```

## Usage

```bash
./swift-deps-diagram \
  --path /path/to/swift/package \
  --mode auto \
  --format png \
  --output deps.png \
  --verbose
```

Flags:
- `--path` package root (default `.`)
- `--project` optional `.xcodeproj` path
- `--workspace` optional `.xcworkspace` path
- `--bazel-targets` optional Bazel query scope (default `//...`)
- `--mode` `auto|spm|xcode|bazel` (default `auto`)
- `--format` `mermaid|dot|png|terminal` (default `png`)
- `--output` output file path (default: stdout for `mermaid`/`dot`/`terminal`, `deps.png` for `png`)
- `--verbose` print generation details for `mermaid`/`dot`/`terminal` file outputs
- `--include-tests` include test targets

Input detection in `auto` mode:
1. Prefer `.xcworkspace` / `.xcodeproj` if found under `--path`
2. Fallback to Bazel workspace markers (`WORKSPACE`, `WORKSPACE.bazel`, `MODULE.bazel`)
3. Fallback to `Package.swift`

## Examples

Sample input projects (SwiftPM and Xcode) are available in `examples/projects/`.

Mermaid only to stdout:

```bash
./swift-deps-diagram --format mermaid
```

DOT only to file:

```bash
./swift-deps-diagram --format dot --output deps.dot
```

ASCII tree to stdout:

```bash
./swift-deps-diagram --format terminal
```

PNG using default output (`deps.png`):

```bash
./swift-deps-diagram --format png
```

PNG mode always prints the absolute output path on stderr.

Verbose message when writing a file:

```bash
./swift-deps-diagram --format dot --output deps.dot --verbose
```

Use explicit Xcode project mode:

```bash
./swift-deps-diagram --mode xcode --project /path/to/App.xcodeproj --format dot --output deps.dot
```

## Using Bazel

Bazel mode reads workspace dependencies using `bazel query` (falls back to `bazelisk` if `bazel` is not found).
If your `bazel` command is Bazelisk, first run may require internet access to download Bazel.

Use explicit Bazel mode for a workspace:

```bash
./swift-deps-diagram --mode bazel --path examples/projects/bazel-basic --format mermaid
```

Limit graph scope to specific Bazel targets:

```bash
./swift-deps-diagram --mode bazel --path examples/projects/bazel-basic --bazel-targets //app:cli --format dot --output deps.dot
```

Include test rules (`*_test`) in Bazel mode:

```bash
./swift-deps-diagram --mode bazel --path examples/projects/bazel-basic --include-tests --format mermaid
```

## Exit Codes

- `0`: success
- `1`: usage/input error (invalid args, missing `Package.swift`)
- `2`: runtime/tooling error (`swift` not found, dump/decode/write failure)
