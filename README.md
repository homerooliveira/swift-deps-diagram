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
  --format both \
  --output deps.txt \
  --png-output deps.png
```

Flags:
- `--path` package root (default `.`)
- `--project` optional `.xcodeproj` path
- `--workspace` optional `.xcworkspace` path
- `--mode` `auto|spm|xcode` (default `auto`)
- `--format` `mermaid|dot|both` (default `both`)
- `--output` output file path (default stdout)
- `--png-output` optional PNG output file generated with Graphviz `dot`
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

Both formats to stdout:

```bash
./swift-deps-diagram --format both
```

Generate Mermaid to stdout and PNG image at the same time:

```bash
./swift-deps-diagram --path ../Bump --format mermaid --png-output bump-deps.png
```

Use explicit Xcode project mode:

```bash
./swift-deps-diagram --mode xcode --project /path/to/App.xcodeproj --format dot --output deps.dot
```

## Exit Codes

- `0`: success
- `1`: usage/input error (invalid args, missing `Package.swift`)
- `2`: runtime/tooling error (`swift` not found, dump/decode/write failure)
