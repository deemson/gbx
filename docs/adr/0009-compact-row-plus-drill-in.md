# 9. Compact main row plus a drill-in detail view

Status: Accepted
Date: 2026-05-22

## Context

Per repo there is more to know than fits one scannable line: per-file diff
stats, the full text of a command error. Cramming it into the row destroys the
at-a-glance overview that is the point of a many-repo table.

## Decision

The main row stays compact: `name | branch (marker if no upstream) | ↑ahead
↓behind | worktree state | [result]`, where worktree state is `clean` or a
changed-path count (with a conflict flag). Anything heavier lives in a
**drill-in**: `enter` on the cursored repo opens a detail screen showing the raw
per-file `+X/-Y` diff vs HEAD and the last command's full error; `esc` returns.

## Consequences

- The table stays scannable; detail is one keystroke away when wanted.
- The drill-in needs a single target, which is why the cursor (ADR 0005) exists.
- Raw `+X/-Y` lines are deliberately *not* in the row.
- The detail diff is loaded on open via `DiffNumStatHead`, off the UI goroutine
  (ADR 0010).
