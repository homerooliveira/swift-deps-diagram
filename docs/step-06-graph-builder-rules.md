# Step 06 - Graph Builder Rules

## Goal
Transform decoded manifest data into a deterministic dependency graph.

## Actions
- Add `internal/graph/build.go` with:
  - `Build(pkg manifest.Package, includeTests bool) (Graph, error)`
- Rules:
  - add all included local targets as nodes
  - target dependency -> edge to local target if present
  - product dependency -> external node ID:
    - `pkg::<package>::<product>` when package exists
    - `product::<product>` otherwise
  - byName dependency:
    - local target match -> local edge
    - else external node `name::<dep>`
- Exclude test targets unless `includeTests == true`.
- Deduplicate edges by `(from,to,kind)`.
- Sort nodes and edges deterministically before returning.

## Deliverables
- full graph construction logic

## Unit Tests
- `TestBuildGraphLocalTargetDependencies`: target-to-target edges are created correctly.
- `TestBuildGraphProductDependencies`: product dependencies create external product nodes.
- `TestBuildGraphByNameLocalResolution`: byName resolves to local target when present.
- `TestBuildGraphByNameExternalFallback`: unresolved byName creates external placeholder node.
- `TestBuildGraphExcludeTestTargetsByDefault`: test targets are excluded when includeTests is false.
- `TestBuildGraphIncludeTestTargetsWhenEnabled`: test targets and edges included when includeTests is true.
- `TestBuildGraphDeduplicatesEdges`: duplicate dependencies collapse to one edge.
- `TestBuildGraphDeterministicOrdering`: repeated builds produce byte-identical node and edge ordering.

## Done Criteria
- builder output is stable across repeated runs.
- fixture-based tests cover all dependency variants.
