# 7. checkout uses a transient prompt + `git switch` guess

Status: Accepted
Date: 2026-05-22

## Context

Unlike `pull`, `checkout` needs an argument: the branch. In an always-focused-
filter UI (ADR 0004) there is no spare text field, and printable keys are taken
by the filter. We also had to choose between `git checkout` and `git switch`.

## Decision

The checkout binding (`ctrl+o`) opens a **transient `branch:` prompt** that
borrows key focus: enter runs, esc cancels. It runs `git switch <branch>` with
guessing left on (the default), so a branch that exists only as a same-named
remote-tracking branch is created locally and set to track it. Failures are
detected by attempt-and-read (ADR 0001) → the row shows `✗`.

## Consequences

- Argument-taking commands have a clear pattern: a transient prompt, not a new
  always-visible field.
- `git switch` is the modern, intent-specific verb; its default guess gives the
  ergonomic "check out the remote branch by name" behaviour for free.
- The prompt is one more `uiMode`; the model already routes keys by mode.
- A future argument-taking command would reuse the same transient-prompt shape.
