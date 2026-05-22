# 2. A fixed, typed command set — view and mutate

Status: Accepted
Date: 2026-05-22

## Context

A multi-repo tool could be a read-only dashboard, or a generic "run this command
everywhere" launcher. The first is safe but limited; the second is powerful but
its output is unstructured and untestable, and it invites destructive misuse.

## Decision

gbx both shows state **and** mutates it, but only through a **fixed set of
commands**, each backed by a typed method on the git wrapper (see ADR 0001). v1
is `pull` and `checkout`. There is no arbitrary command passthrough.

## Consequences

- Each command's output is structured, so results render inline and are tested.
- The blast radius is bounded: only operations we have deliberately added exist.
- New cross-repo operations are a deliberate act (wrapper method + binding +
  tests), not a free-form text box.
- Power users wanting arbitrary git will find gbx insufficient — accepted; that
  is what a shell loop is for. See ADR 0006 for the no-confirmation stance that
  this bounded set makes acceptable.
