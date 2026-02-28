# Implementation Guide for Parity Port

This guide describes how to reimplement the system in another language while preserving behavior parity.

## 1. Target Architecture Blueprint

Mirror the architecture as language-neutral components.

| Port module | Responsibility |
|---|---|
| `cli` | Parse flags, validate constraints, print warnings/errors, map failures to exit codes |
| `app` | Orchestrate resolve -> load -> build graph -> render -> output |
| `resolver` | Resolve user input into a normalized execution plan |
| `adapter_swiftpm` | Execute SwiftPM manifest extraction command with timeout |
| `adapter_xcode` | Load and normalize Xcode project/workspace dependency data |
| `adapter_tuist` | Generate Xcode project from `Project.swift` (`tuist generate --no-open`) |
| `adapter_bazel` | Load and normalize Bazel dependency data via query commands |
| `graph_core` | Canonical graph types, edge keying, sorting helpers |
| `graph_from_spm` | Convert SwiftPM dependency model into canonical graph |
| `graph_from_xcode` | Convert Xcode dependency model into canonical graph |
| `graph_from_bazel` | Convert Bazel dependency model into canonical graph |
| `render_mermaid` | Convert canonical graph to Mermaid text |
| `render_dot` | Convert canonical graph to DOT text |
| `render_terminal` | Convert canonical graph to terminal ASCII tree text |
| `output_text` | Write text output to stdout or atomically to file |
| `output_png` | Convert DOT to PNG through Graphviz |
| `errors` | Typed failure categories and exit-code mapping |

## 2. Porting Order (Phased)

### Phase 1: errors + graph model

Implement first:
- Typed failure kinds.
- Error wrapper with optional cause.
- Exit code mapping.
- Canonical graph model and deterministic sorting.
- Edge deduplication key function.

Definition of done:
- Error-class mapping and graph ordering tests pass.

### Phase 2: resolver

Implement:
- Mode validation (`auto|spm|xcode|bazel`).
- Path normalization and mutually exclusive flag handling.
- `auto` precedence and workspace/project tie-breakers.

Definition of done:
- Resolver behavior matches precedence/error contract.

### Phase 3: SwiftPM path

Implement:
- Manifest decode model with dependency-kind normalization.
- SwiftPM command integration with timeout and stderr propagation.
- SwiftPM-to-graph conversion.

Definition of done:
- Dependency-kind handling and test filtering parity achieved.

### Phase 4: render/output

Implement:
- Mermaid renderer and escaping rules.
- DOT renderer, style mapping, and top-to-bottom orientation.
- Text output writer with atomic write semantics.

Definition of done:
- Golden-like output and determinism tests pass.

### Phase 5: Xcode/Bazel

Implement:
- Xcode adapter + graph conversion.
- Bazel adapter + graph conversion.
- Include-tests behavior across both adapters.

Definition of done:
- Adapter conversion and filtering parity tests pass.

### Phase 6: PNG integration + hardening

Implement:
- DOT-to-PNG Graphviz integration with timeout.
- Final orchestration and CLI integration.
- Error/warning/logging contracts and end-to-end matrix tests.

Definition of done:
- All mode/format scenarios satisfy parity matrix and completion checklist.

## 3. Normative Pseudocode

### 3.1 End-to-end `Run` orchestration

```text
function Run(ctx, opts, stdout):
  validateOptions(opts)

  resolved = resolveInput({
    path: opts.path,
    mode: opts.mode,
    projectPath: opts.projectPath,
    workspacePath: opts.workspacePath,
    bazelTargets: opts.bazelTargets,
  })

  switch resolved.mode:
    case SPM:
      manifestBytes = dumpPackage(ctx, resolved.packagePath)
      manifest = decodeManifest(manifestBytes)
      graph = buildGraphFromSPM(manifest, opts.includeTests)

    case XCODE:
      if resolved.tuistPath != "":
        generateTuistProject(ctx, resolved.tuistPath)
        generated = resolveInput({path: resolved.tuistPath, mode: "xcode"})
        if generated.projectPath == "":
          raise runtime_failed("tuist generation completed but no xcode project was resolved")
        resolved = generated
      project = loadXcodeProject(ctx, resolved.projectPath)
      graph = buildGraphFromXcode(project, opts.includeTests)

    case BAZEL:
      workspace = loadBazelWorkspace(ctx, resolved.workspacePath, resolved.targets)
      graph = buildGraphFromBazel(workspace, opts.includeTests)

    default:
      raise invalid_args

  if opts.format == "png":
    dotText = renderDot(graph)
    outputPath = opts.outputPath if opts.outputPath != "" else "deps.png"
    writePng(ctx, dotText, outputPath)
    logStderr("generated png using dot format at " + outputPath)
    return

  if opts.format == "mermaid":
    text = renderMermaid(graph)
  else if opts.format == "dot":
    text = renderDot(graph)
  else:
    text = renderTerminal(graph)
  writeText(text, opts.outputPath, stdout)

  if opts.verbose and opts.outputPath != "":
    if opts.format == "mermaid":
      logStderr("generated mermaid content at " + opts.outputPath)
    else if opts.format == "dot":
      logStderr("generated dot content at " + opts.outputPath)
    else if opts.format == "terminal":
      logStderr("generated terminal content at " + opts.outputPath)
```

### 3.2 Resolver algorithm (`auto|spm|xcode|bazel`)

```text
function Resolve(request):
  path = absolute(request.path or ".")
  mode = request.mode or "auto"

  reject if mode invalid
  reject if both projectPath and workspacePath set

  if mode == "spm":
    return {mode:spm, packagePath:resolvePackagePath(path)}

  if mode == "xcode":
    if tryResolveXcode(path, request.projectPath, request.workspacePath) succeeds:
      return xcode result
    if request.projectPath or request.workspacePath:
      rethrow xcode error
    tuistPath = resolveTuistPath(path)  # detects Project.swift
    return {mode:xcode, tuistPath}

  if mode == "bazel":
    workspacePath = resolveBazelWorkspacePath(path)
    return {mode:bazel, workspacePath, targets:normalizeTargets(request.bazelTargets)}

  # auto mode
  if request.projectPath or request.workspacePath:
    projectPath, workspacePath = resolveXcodePath(path, request.projectPath, request.workspacePath)
    return {mode:xcode, projectPath, workspacePath}

  if tryResolveXcode(path) succeeds:
    return xcode result

  if tryResolveTuist(path) succeeds:
    return {mode:xcode, tuistPath}

  if tryResolveBazel(path) succeeds:
    return bazel result
  if bazel failed for reason other than missing marker:
    rethrow

  if tryResolveSpm(path) succeeds:
    return spm result
  if spm failed for reason other than missing manifest:
    rethrow

  raise input_not_found with combined marker-check message
```

### 3.3 Graph build algorithms

#### SPM graph builder

```text
initialize nodes, edges, edgeDedup, localTargets

for target in manifest.targets:
  if not includeTests and target.type == "test": continue
  localTargets.add(target.name)
  add target node

for target in manifest.targets:
  if not includeTests and target.type == "test": continue
  from = target node ID

  for dep in target.dependencies:
    switch dep.kind:
      target:
        if dep.name empty: continue
        to = local target ID if dep.name in localTargets else external product fallback ID
        add edge kind=target

      product:
        if dep.name empty: continue
        to = pkg-based product ID or product-only ID
        add external node
        add edge kind=product

      by_name:
        if dep.name empty: continue
        to = local target ID if dep.name in localTargets else name-based fallback ID
        add fallback node when external
        add edge kind=by_name

      default:
        continue

dedupe edges by (kind, from, to)
sort edges deterministically
```

#### Xcode graph builder

```text
sort targets by (name, id)
create targetID->nodeID map, resolving duplicate names with suffix
exclude test targets when includeTests=false

for each included target:
  add target node

for each included target:
  for target dependency in targetDependsOn:
    if dependency target exists in map:
      add edge kind=target

  for package product dependency:
    if product name empty: continue
    add external product node (identity-aware when available)
    add edge kind=product

dedupe and sort edges
```

#### Bazel graph builder

```text
sort workspace targets by label
exclude _test rules when includeTests=false
add included local target nodes

for each included target:
  for dep in sorted(deps):
    if dep starts with "@":
      add external node
      add edge kind=product
    else if dep starts with "//":
      if dep is known test rule and tests excluded: continue
      add local target node if absent
      add edge kind=target
    else:
      ignore

dedupe and sort edges
```

### 3.4 Render + output + error mapping flow

```text
if format in {mermaid, dot, terminal}:
  text = render(format, graph)
  write text to stdout when output path empty, otherwise atomic file write

if format == png:
  dotText = render(dot, graph)
  write PNG through Graphviz

on failure:
  return typed failure category
  CLI maps category to process exit code (1 or 2)
```

## 4. Interfaces and Contracts for Reimplementation

### 4.1 Command runner / process execution

```text
interface CommandRunner {
  lookPath(binaryName) -> (resolvedPath, error)
  run(context, workingDirectory, executable, args[]) -> (stdoutBytes, stderrBytes, error)
}
```

Contract:
- Must preserve stderr text for user-facing failure messages.
- Must support cancellation/timeouts.
- Must support explicit working directory.

### 4.2 Filesystem abstraction

```text
interface FileSystem {
  stat(path) -> (info, error)
  readFile(path) -> (bytes, error)
  mkdirAll(path, mode) -> error
  createTemp(dir, pattern) -> (fileHandle, error)
  rename(src, dst) -> error
  glob(pattern) -> (paths[], error)
  absolute(path) -> (path, error)
}
```

Contract:
- Resolver depends on directory/file distinction and path normalization.
- Text output requires same-directory temp creation + rename.

### 4.3 Clock/timeouts

```text
interface TimeoutProvider {
  withTimeout(parentContext, duration) -> (childContext, cancel)
}
```

Required timeout policy:
- SwiftPM manifest extraction: 30s
- Xcode conversion/parsing command: 30s
- Graphviz PNG render: 30s
- Bazel query command: 2m per query invocation

### 4.4 Renderer abstraction

```text
interface Renderer {
  mermaid(graph) -> (text, error)
  dot(graph) -> (text, error)
  terminal(graph) -> (text, error)
}
```

Contract:
- Must be deterministic for identical logical graph input.
- Must fail on invalid edge references when rendering.

### 4.5 I/O and error contracts

| Contract area | Input | Output | Required behavior |
|---|---|---|---|
| Resolver | user request | normalized resolution | preserves precedence and typed not-found classes |
| Adapters | resolved paths/scope | normalized source model | uses external tools with timeout + stderr propagation |
| Graph builders | normalized model + test flag | canonical graph | deduped, deterministically sorted edges |
| Renderers | canonical graph | Mermaid/DOT/terminal text | exact structure + escaping contract |
| Output writers | content + target path | persisted output | stdout or atomic file write |
| CLI error bridge | typed failures | process code + stderr text | exact `0/1/2` mapping |

## 5. Behavior-Parity Matrix

| Required behavior | Expected observable result |
|---|---|
| Invalid mode/format rejected | Error text + exit code 1 |
| Project/workspace mutual exclusion enforced | Error + exit code 1 |
| Positional args rejected | Error + exit code 1 |
| SPM mode with Xcode flags emits warning | Warning on stderr |
| `auto` precedence is Xcode > Bazel > SwiftPM | Selected mode follows precedence |
| Empty Bazel scope normalizes to `//...` | Scope used as `//...` |
| Missing Swift tool is explicit failure class | Tool-not-found failure |
| Swift command failures include stderr text | Rich failure message |
| Xcode loader extracts target + package product deps | Canonical model includes both |
| Bazel queries include noimplicit/notool flags | Query behavior filtered as specified |
| Bazel failure includes stderr text | Rich failure message |
| Edge dedup key is `(kind,from,to)` | No duplicate edges |
| SPM byName local vs external split | Local maps to target; external uses `name::` |
| Test filtering defaults to exclude | Test nodes/edges absent unless enabled |
| Mermaid output contract | `flowchart TD`, deterministic aliasing, escaping |
| DOT output contract | `digraph dependencies`, `rankdir=TB`, expected styles/escaping |
| Terminal output contract | ASCII tree roots, deterministic traversal, per-branch expansion, cycle marker `(*)` |
| PNG default output path | `deps.png` when no output path provided |
| PNG success log message | Message appears on success |
| Text output file write is atomic pattern | Complete file produced via temp+rename |
| Error category to exit mapping | Exact `1` vs `2` semantics |

Include parity verification for edge cases:
- Missing external tools (`swift`, `plutil`, `bazel`/`bazelisk`, `dot`)
- Malformed manifest/query output
- Unknown dependency variants (ignored)
- Duplicate edges in input
- Include/exclude tests across all adapters

## 6. Conformance Test Plan for the New Language

### 6.1 Unit tests by module

- `errors`: failure-category mapping and exit code conversion.
- `resolver`: precedence, explicit-path behavior, marker detection, not-found aggregation.
- `swiftpm adapter`: command invocation, timeout, stderr propagation.
- `xcode adapter`: extraction of targets/target deps/product deps.
- `bazel adapter`: query construction, filter flags, parse/error behavior.
- `graph builders`: ID conventions, edge kinds, dedupe, include-tests behavior.
- `renderers`: output structure, escaping, determinism.
- `output`: stdout and atomic file-write behavior.

### 6.2 Integration test scenarios (mode/format matrix)

Matrix dimensions:
- Mode: `auto`, `spm`, `xcode`, `bazel`
- Format: `mermaid`, `dot`, `png`, `terminal`
- Include-tests: on/off
- Output: stdout and explicit file path

Assertions per run:
- Exit code
- stderr/stdout contract
- Output artifact path/content

### 6.3 Golden output tests

- Golden Mermaid output for fixed graphs.
- Golden DOT output for fixed graphs including `rankdir=TB`.
- Golden terminal ASCII tree output for fixed graphs.
- Repeat-run determinism checks with shuffled input insertion order.

### 6.4 Exit code and message contract tests

- Input/argument class failures produce exit code `1`.
- Runtime/tool/parse/render/output failures produce exit code `2`.
- Tool stderr is surfaced where specified.

## 7. Porting Pitfalls and Mitigations

| Pitfall | Impact | Mitigation |
|---|---|---|
| Non-deterministic iteration of hash/map collections | Output drift across runs | Always sort nodes/edges before rendering |
| Platform path handling differences | Resolver mismatch | Normalize paths early and consistently |
| Dropping stderr details from tool calls | Weak diagnostics and contract mismatch | Preserve stderr in typed errors |
| Timeout model differences | Hung commands or premature cancellation | Use per-command timeout policy exactly |
| Non-atomic file write implementation | Partial/corrupt outputs | Use temp-in-destination-dir + rename |
| Incorrect DOT escaping | Invalid DOT or changed text output | Apply exact escaping transform order |
| Incorrect Mermaid escaping | Changed labels/output mismatch | Apply exact escaping transform order |
| Workspace project selection shortcuts | Wrong Xcode project resolution | Honor workspace references first, sorted fallback second |
| Bazel query scope drift | Extra/missing graph edges | Keep rule/deps query expressions and filters exact |

## 8. Completion Criteria

A port is equivalent only when all checks pass:

- [ ] CLI contract matches flags/defaults/validation behavior.
- [ ] Resolver precedence and tie-breakers match exactly.
- [ ] SwiftPM, Xcode, and Bazel adapters produce equivalent normalized models.
- [ ] Canonical graph IDs, kinds, dedup, and sorting behavior match.
- [ ] Mermaid and DOT outputs match golden expectations.
- [ ] PNG generation path and failure handling match.
- [ ] Verbose/non-verbose logging behavior matches output rules.
- [ ] Error categories and `0/1/2` exit mapping match.
- [ ] Determinism tests pass over repeated runs.
- [ ] All parity-matrix scenarios pass, including edge cases.
