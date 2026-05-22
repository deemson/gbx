# 1. All git access goes through a tested wrapper

Status: Accepted
Date: 2026-05-22

## Context

gbx is, at heart, a fan-out over `git`. If git invocations were scattered across
the UI and elsewhere, they would be untestable (string-built command lines,
ad-hoc output parsing) and inconsistent in how failures surface.

## Decision

All git access goes through `internal/git`: each operation is a typed method on
`Repo` (e.g. `Status`, `DiffNumStatHead`, `Pull`, `Switch`) returning structured
data or a typed error. The raw process runner lives in `internal/git/exec`; test
fixtures that build real repos live in `internal/git/gitest`. Nothing outside
`internal/git` shells out to `git`.

Errors are mapped by **attempt-and-read**: run the command, then inspect exit
code + stderr to produce a typed error (a specific sentinel where we recognise
the failure, `UnknownRunError` otherwise). We do not pre-flight with extra git
calls to predict failures.

## Consequences

- Every git capability is exercised by tests against real repositories.
- Adding a command means adding one typed method + its test — a uniform path.
- Output parsing is centralised and reusable (the TUI has no git knowledge).
- Attempt-and-read keeps the happy path to a single subprocess and makes the
  error taxonomy grow only as fast as we actually handle distinct failures.
