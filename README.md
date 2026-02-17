# swift-deps-diagram

CLI tool to generate dependency diagrams from a Swift Package manifest (`Package.swift`) by using `swift package dump-package`.

## Build

```bash
go build ./cmd/swift-deps-diagram
```

## Usage

```bash
./swift-deps-diagram \
  --path /path/to/swift/package \
  --format both \
  --output deps.txt
```

Flags:
- `--path` package root (default `.`)
- `--format` `mermaid|dot|both` (default `both`)
- `--output` output file path (default stdout)
- `--include-tests` include test targets

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

## Exit Codes

- `0`: success
- `1`: usage/input error (invalid args, missing `Package.swift`)
- `2`: runtime/tooling error (`swift` not found, dump/decode/write failure)
