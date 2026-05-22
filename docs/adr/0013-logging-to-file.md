# 13. Logging to `~/gbx.log`; the TUI owns stdout

Status: Accepted
Date: 2026-05-22

## Context

A full-screen alt-screen TUI owns the terminal: anything written to stdout/
stderr corrupts the rendered frame. But we still need diagnostics, especially
for the off-goroutine git commands whose failures would otherwise be invisible.

## Decision

Log with zerolog to a file (`~/gbx.log`); the TUI keeps exclusive ownership of
stdout. Off-goroutine commands log their own failures (e.g. a failed `pull`)
before returning a result message. Tests discard logs (wired in `TestMain`).

## Consequences

- The rendered UI is never corrupted by stray log lines.
- There is a durable record of what git did, separate from the on-screen `✗`,
  for diagnosing failures after the fact.
- Logs are out of band: a user must `tail ~/gbx.log` to see them; the in-app
  surface for an error is the drill-in (ADR 0009).
