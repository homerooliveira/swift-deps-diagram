# swift-deps-diagram

CLI tool to generate dependency diagrams from a Swift Package manifest (`Package.swift`) by using `swift package dump-package`.
Also supports Xcode projects/workspaces that use Swift Package Manager dependencies.

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
- `--mode` `auto|spm|xcode` (default `auto`)
- `--format` `mermaid|dot|png` (default `png`)
- `--output` output file path (default stdout)
- `--verbose` print generation details for file outputs
- `--include-tests` include test targets

Input detection in `auto` mode:
1. Prefer `.xcworkspace` / `.xcodeproj` if found under `--path`
2. Fallback to `Package.swift`

## Examples

Mermaid only to stdout:

```bash
./swift-deps-diagram --format mermaid
```

DOT only to file:

```bash
./swift-deps-diagram --format dot --output deps.dot
```

PNG using default output (`deps.png`):

```bash
./swift-deps-diagram --format png
```

Verbose message when writing a file:

```bash
./swift-deps-diagram --format dot --output deps.dot --verbose
```

Use explicit Xcode project mode:

```bash
./swift-deps-diagram --mode xcode --project /path/to/App.xcodeproj --format dot --output deps.dot
```

## Exit Codes

- `0`: success
- `1`: usage/input error (invalid args, missing `Package.swift`)
- `2`: runtime/tooling error (`swift` not found, dump/decode/write failure)
