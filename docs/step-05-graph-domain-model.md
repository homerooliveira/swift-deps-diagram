# Step 05 - Dependency Graph Domain Model

## Goal
Define graph primitives used by builder and renderers.

## Actions
- Add `internal/graph/model.go`:
  - `NodeKind`: `target`, `external_product`
  - `EdgeKind`: `target`, `product`, `by_name`
  - `Node { ID, Label, Kind }`
  - `Edge { FromID, ToID, Kind }`
  - `Graph { Nodes map[string]Node, Edges []Edge }`
- Add helpers for deterministic ordering and edge dedup keys.

## Deliverables
- stable internal graph API for remaining steps

## Unit Tests
- `TestNodeKindValuesStable`: expected enum/string values remain unchanged.
- `TestEdgeKindValuesStable`: expected edge kind values remain unchanged.
- `TestGraphEdgeDedupKey`: identical edge tuples generate identical dedup keys.
- `TestGraphSortOrder`: node and edge ordering helpers produce deterministic order.

## Done Criteria
- graph model compiles and can represent local + external dependencies.
