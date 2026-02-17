# Step 08 - Graphviz DOT Renderer

## Goal
Render the same graph to DOT format for Graphviz tooling.

## Actions
- Add `internal/render/dot.go`:
  - `Dot(g graph.Graph) (string, error)`
- Emit:
  - `digraph dependencies { ... }`
  - quoted node IDs and labels
  - edges in deterministic order
- Apply style by kind:
  - target: `shape=box`
  - external_product: `shape=ellipse,style=dashed`

## Deliverables
- DOT renderer aligned with graph model semantics

## Unit Tests
- `TestDotIncludesGraphHeaderAndFooter`: output contains valid digraph wrapper.
- `TestDotRendersAllNodesAndEdges`: fixture graph renders expected node and edge lines.
- `TestDotQuotesIDsAndLabels`: IDs and labels are quoted and escaped correctly.
- `TestDotAppliesNodeStylesByKind`: target and external nodes receive expected style attributes.
- `TestDotDeterministicOutput`: repeated renders produce byte-identical output.

## Done Criteria
- output can be consumed by `dot` CLI without syntax errors.
