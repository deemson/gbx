# 8. Inline results, per-row auto-refresh, manual `ctrl+r`

Status: Accepted
Date: 2026-05-22

## Context

Running a command across repos produces per-repo outcomes and changes per-repo
git state. We could surface results in a separate log pane, and we could keep
state fresh with background polling. Both add UI and cost.

## Decision

- **Results are inline.** Each row has a result cell showing a spinner while a
  command runs, then `✓` or `✗`. There is no separate log pane.
- **A row auto-refreshes** its git state after its command completes.
- **`ctrl+r`** triggers a manual refresh of the filtered repos.
- **No periodic background refresh** in v1.

## Consequences

- The table is the single surface: state and outcomes live in one place.
- After a `pull`, ahead/behind and worktree state update on their own.
- State can be stale relative to changes made outside gbx until `ctrl+r`;
  accepted, and cheap to refresh on demand.
- Detailed failure text does not fit a one-character cell — it lives in the
  drill-in (ADR 0009).
