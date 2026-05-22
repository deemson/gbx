# 10. Concurrency: one `tea.Cmd` per repo, message-driven

Status: Accepted
Date: 2026-05-22

## Context

Every git operation is a subprocess that blocks for tens to hundreds of
milliseconds. Doing them serially on the UI goroutine would freeze the interface
and make a 30-repo `pull` feel broken.

## Decision

Repos are opened and queried **concurrently**: one `tea.Cmd` per repo per
operation, run off the UI goroutine. Results come back as messages
(`repoFoundMsg`, `statusLoadedMsg`, `cmdDoneMsg`, `detailLoadedMsg`, …) that
update individual rows in `Update`. The model is updated only on the UI
goroutine, by message; commands never touch model state directly.

## Consequences

- The UI stays responsive; rows fill in and update independently as work lands.
- All concurrency is funnelled through Bubble Tea's command/message model — no
  shared mutable state, no manual locking in the model.
- Ordering is not guaranteed; handlers are written to tolerate results arriving
  in any order and for repos that may have been filtered away.
