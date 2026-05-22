# 4. fzf-style: always-focused filter, non-printable actions

Status: Accepted
Date: 2026-05-22

## Context

With many repos on screen, selecting targets is the core interaction. A modal
design (a "filter mode" you enter and leave) adds state and keystrokes. fzf
showed that an always-on filter is faster and more intuitive.

## Decision

A filter text input is **always focused** from startup. **Printable keys edit
the filter**; **every action is a non-printable binding** (`ctrl+…`, arrows,
enter, esc). The filter fuzzy-matches repo names. Transient sub-screens (the
checkout `branch:` prompt, the drill-in, the help overlay) capture keys while
open and hand focus back on close.

Action bindings avoid terminal-reserved combos: `ctrl+c/z/s/q`, and
`ctrl+m/i/h/[` (which are enter/tab/backspace/esc).

## Consequences

- Filtering is immediate and modeless — no "press / to search".
- The set of possible actions is constrained to non-printable keys, which keeps
  the keymap small and predictable (see KEYMAP.md / the `ctrl+g` overlay).
- A textual command like `?` for help is impossible by construction; help is a
  binding (`ctrl+g`) instead.
- Mode handling (`uiMode` in the model) is needed so sub-screens can borrow key
  focus from the filter.
