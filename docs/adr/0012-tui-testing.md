# 12. TUI tested end-to-end, with model-level tests for in-place changes

Status: Accepted
Date: 2026-05-22

## Context

A TUI is easy to leave untested ("it's just rendering"). But gbx's behaviour —
filtering, fan-out commands, mode transitions, result glyphs — is logic worth
testing. The challenge: the alt-screen renderer does differential,
cursor-positioned updates, so an *in-place* change (e.g. `↓1`→`↓0`, the cursor
moving) is not a contiguous substring of the output stream.

## Decision

Test the TUI **end-to-end** with the `testProgram` harness (`internal/tui`,
`testhelper_test.go`): it drives a real `tea.Program`, injects keys with
`send`/`sendKey`, and asserts rendered output with `waitForContent`. Fixtures
are real repos built with `gitest`.

Because `waitForContent` only reliably sees *fresh/appended* text, assert state
**transitions** with renderer-free **model-level tests** that drive
`model.Update` directly and inspect state (`model_test.go`).

## Consequences

- Behaviour is covered at two levels: real-program output for what the user
  sees appear, and model state for transitions the renderer hides.
- New features come with both kinds of test where relevant (e.g. drill-in: a
  model test for the mode switch, an e2e test for the detail text appearing).
- Tests use real git repositories, so they catch wrapper/TUI integration bugs.
