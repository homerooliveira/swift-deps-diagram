# Example Projects

This folder contains small projects you can use to test the CLI.

## `hello-spm`

Single-package executable that depends on Swift package product `Alamofire`.

Run:

```bash
go run ./cmd/swift-deps-diagram --path examples/projects/hello-spm --mode spm --format mermaid
```

## `local-path-deps`

Multi-package example that uses local path dependencies:

- `App` depends on `FeatureKit`
- `App` depends on Swift package product `Alamofire`
- `FeatureKit` depends on `CoreKit`

Run for `App`:

```bash
go run ./cmd/swift-deps-diagram --path examples/projects/local-path-deps/App --mode spm --format mermaid
```

Run for `FeatureKit`:

```bash
go run ./cmd/swift-deps-diagram --path examples/projects/local-path-deps/FeatureKit --mode spm --format mermaid
```

## `xcodeproj-basic`

Minimal `.xcodeproj` example with:

- `App` target depending on `Core` target
- `App` target depending on Swift package product `Alamofire`

Run:

```bash
go run ./cmd/swift-deps-diagram --mode xcode --project examples/projects/xcodeproj-basic/App.xcodeproj --format mermaid
```

## `xcworkspace-basic`

Minimal `.xcworkspace` example that references `App.xcodeproj` and includes a Swift package product dependency (`SnapshotTesting`).

Run:

```bash
go run ./cmd/swift-deps-diagram --mode xcode --workspace examples/projects/xcworkspace-basic/App.xcworkspace --format mermaid
```

Note: Xcode mode requires `plutil` (macOS/Xcode command-line tooling).
