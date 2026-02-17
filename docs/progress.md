# Implementation Progress

Use this file to track completion of major documentation and implementation milestones.

Status legend:
- `TODO`: not started
- `IN_PROGRESS`: currently being implemented
- `DONE`: implemented and verified
- `BLOCKED`: waiting on dependency/decision

## Overall Status

- Current status: `DONE`
- Last updated: `2026-02-17`

## Milestones

| Milestone | Status | Owner | Notes |
|---|---|---|---|
| Go CLI foundation | DONE | Codex | CLI, graph model, renderers, output |
| SwiftPM dependency graph support | DONE | Codex | `dump-package` integration and tests |
| PNG generation support | DONE | Codex | Graphviz integration |
| Xcode + SPM dependency graph support | DONE | Codex | auto detection, `.pbxproj` parsing |

## Update Rules

1. Mark only active work as `IN_PROGRESS`.
2. Move a milestone to `DONE` only after verification is complete.
3. Keep `Last updated` current when any status changes.
4. Add blockers and decision notes in `Notes`.
