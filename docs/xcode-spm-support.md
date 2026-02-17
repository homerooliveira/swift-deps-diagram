# Xcode + SPM Support

This project now supports dependency graph generation from:
- Swift Package manifests (`Package.swift`)
- Xcode projects/workspaces (`.xcodeproj` / `.xcworkspace`) that use Swift Package dependencies

## Input Modes

- `--mode auto` (default):
  1. Prefer Xcode input if `.xcworkspace` or `.xcodeproj` exists under `--path`
  2. Fallback to `Package.swift`
- `--mode spm`: force Swift Package manifest mode
- `--mode xcode`: force Xcode mode

## Xcode-Specific Flags

- `--project /path/to/App.xcodeproj`
- `--workspace /path/to/App.xcworkspace`

Notes:
- `--project` and `--workspace` are mutually exclusive.
- In workspace mode, the tool resolves the underlying `.xcodeproj` and parses `project.pbxproj`.

## What Xcode Mode Graphs

- Xcode targets
- Target-to-target dependencies
- Target-to-Swift-Package product dependencies

## Examples

Auto detect from current path:

```bash
./swift-deps-diagram --path . --format png
```

Force Xcode mode using a project:

```bash
./swift-deps-diagram --mode xcode --project /path/to/App.xcodeproj --format dot --output deps.dot
```

Force SwiftPM mode even when Xcode files exist:

```bash
./swift-deps-diagram --mode spm --path /path/to/repo --format mermaid --output deps.mmd
```

Generate PNG from Xcode graph:

```bash
./swift-deps-diagram --mode xcode --project /path/to/App.xcodeproj --format png --output deps.png --verbose
```
