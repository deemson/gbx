# 6. Commands fire immediately, without confirmation

Status: Accepted
Date: 2026-05-22

## Context

Cross-repo mutation is exactly where an accidental keystroke could do wide
damage, which argues for a confirmation step. But confirmation prompts on every
action are friction, and gbx already has structural guards.

## Decision

Commands fire **immediately** on their binding — no "are you sure?" step.

This is acceptable because the surrounding decisions bound the risk: the command
set is fixed and non-destructive in v1 (`pull`, `checkout` — see ADR 0002), the
target set is exactly the visible filtered rows (ADR 0005), and each result is
shown inline so mistakes are immediately apparent (ADR 0008).

## Consequences

- Fast: filter, press a binding, done.
- The safety story rests on the command set staying non-destructive. **A command
  that can lose work (e.g. a future `reset --hard`, `clean`, force-push) must
  revisit this ADR** and likely introduce confirmation for that command.
- Security review is warranted whenever a mutating command is added
  (see ADR 0011).
