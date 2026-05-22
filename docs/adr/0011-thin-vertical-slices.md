# 11. Build in thin vertical slices; reviews on request

Status: Accepted
Date: 2026-05-22

## Context

The product was specified requirements-light. Building it as horizontal layers
(all of the git wrapper, then all of the TUI) would defer any working,
demonstrable behaviour to the very end and hide integration risk.

## Decision

Build in **thin vertical slices**: each slice is one visible capability
end-to-end (git wrapper method + TUI wiring + tests with `gitest` and the
`testProgram` harness), coded and tested and then **committed before the next**.
The roadmap in `DECISIONS.md` lists the slices. Code review and security review
are run **on request**, not automatically — security review especially when a
mutating command is added (see ADR 0006).

## Consequences

- Every commit leaves a working, tested program.
- Integration problems surface in the slice that introduces them, not at the
  end.
- The git history reads as a sequence of capabilities (`slice N: …`).
- Reviews being on-request keeps iteration fast but puts the onus on us to ask
  for security review at the right moments.
