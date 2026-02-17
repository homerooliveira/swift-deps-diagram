# Step 07 - Mermaid Renderer

## Goal
Render graph data to Mermaid syntax suitable for Markdown documentation.

## Actions
- Add `internal/render/mermaid.go`:
  - `Mermaid(g graph.Graph) (string, error)`
- Emit:
  - `flowchart TD`
  - node declarations with readable labels
  - `A --> B` edges in deterministic order
- Use renderer-local IDs (`n1`, `n2`, ...) mapped from sorted nodes to avoid escaping issues.
- Escape labels for Mermaid safety.

## Deliverables
- Mermaid renderer with deterministic output

## Unit Tests
- `TestMermaidIncludesFlowchartHeader`: output starts with `flowchart TD`.
- `TestMermaidRendersAllNodesAndEdges`: known graph fixture renders expected relationships.
- `TestMermaidEscapesLabels`: labels with special characters are safely escaped.
- `TestMermaidUsesDeterministicNodeIDs`: renderer-local IDs remain deterministic for same input.
- `TestMermaidDeterministicOutput`: repeated renders produce byte-identical output.

## Done Criteria
- output validates in Mermaid-compatible viewers.
- same input graph always produces byte-identical output.
