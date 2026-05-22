---
name: charm-tui
description: Local Bubble Tea v2, Bubbles v2, and Lip Gloss v2 examples and tutorials matching this project's exact stack. Use when writing or editing any TUI code that imports charm.land/bubbletea/v2, charm.land/bubbles/v2, or charm.land/lipgloss/v2 — Model/Init/Update/View, tea.Cmd, key bindings, list/table/textinput/help/spinner/progress/pager bubbles, or Lip Gloss styles, layout, tables, lists, trees, and color.
---

# charm-tui

Bundled, **v2** Charm examples matching the project's stack
(`bubbletea/v2`, `bubbles/v2`, `lipgloss/v2`). Read the relevant `main.go`
before writing TUI code — **do not** rely on recalled API shape, which is
mostly v1 and wrong (e.g. v2 imports are `charm.land/...v2`, and `Update` /
`View` signatures and message types differ from v1).

## How to use

1. Find the closest example/tutorial in the maps below.
2. `Read` its `main.go` (and `README.md` when present) to copy the v2 idiom.
3. Match the example's patterns, then adapt.

Full indexes: `examples/bubbletea/examples/README.md` and the `lipgloss`
example dirs. Paths below are relative to this skill directory.

## Learn the fundamentals first

- `examples/bubbletea/tutorials/basics/` — Model/Init/Update/View, key
  handling, the core loop.
- `examples/bubbletea/tutorials/commands/` — `tea.Cmd`, async work, custom
  messages.

## Bubble Tea — by need (`examples/bubbletea/examples/<dir>/main.go`)

| Need | Example dir |
|---|---|
| Filter / text input | `textinput`, `textinputs` |
| Scrollable list + filtering | `list-simple`, `list-default`, `list-fancy` |
| Tabular rows | `table`, `table-resize` |
| Scrollable detail / viewport | `pager` |
| Switch views / drill-in | `views`, `composable-views`, `tabs` |
| Help overlay | `help` |
| Spinner / loading | `spinner`, `spinners` |
| Progress bar | `progress-animated`, `progress-static`, `progress-download` |
| Async / streamed messages | `realtime`, `send-msg`, `http`, `debounce` |
| Long-running work + TUI | `tui-daemon-combo`, `package-manager` |
| Run external command (e.g. git) | `exec` |
| Sequence/batch commands | `sequence` |
| Return a final value on quit | `result` |
| Alt screen / window size | `altscreen-toggle`, `fullscreen`, `window-size` |
| Focus, key enhancements | `focus-blur`, `keyboard-enhancements`, `print-key` |

## Lip Gloss — by need (`examples/lipgloss/examples/<dir>/`)

| Need | Example dir |
|---|---|
| Styles, borders, layout | `layout` |
| Styled tables | `table/` (e.g. `table/languages`) |
| Styled lists | `list/` (e.g. `list/simple`) |
| Trees | `tree/` (e.g. `tree/simple`) |
| Color / adaptive color | `color`, `compat` |
| Gradients / blending | `blending/`, `brightness` |
| Free-form drawing | `canvas` |
