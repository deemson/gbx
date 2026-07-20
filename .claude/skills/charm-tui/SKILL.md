---
name: charm-tui
description: Local Bubble Tea v2, Bubbles v2, and Lip Gloss v2 examples and tutorials matching this project's exact stack. Use when writing or editing any TUI code that imports charm.land/bubbletea/v2, charm.land/bubbles/v2, or charm.land/lipgloss/v2 — Model/Init/Update/View, tea.Cmd, key bindings, list/table/textinput/help/spinner/progress/pager bubbles, or Lip Gloss styles, layout, tables, lists, trees, and color.
---

# charm-tui

**v2** Charm examples matching the project's stack (`bubbletea/v2`,
`bubbles/v2`, `lipgloss/v2`). Read the relevant `main.go` before writing TUI
code — **do not** rely on recalled API shape, which is mostly v1 and wrong
(e.g. v2 imports are `charm.land/...v2`, and `Update` / `View` signatures and
message types differ from v1).

## Getting the repos

The examples are nested modules, so they are **not** in the module cache —
they only exist in a clone. Clone each repo at the tag matching `go.mod`:

```sh
v=$(go list -m -f '{{.Version}}' charm.land/bubbletea/v2)
[ -d ".charm-repos/bubbletea@$v" ] || git -c advice.detachedHead=false clone \
  --quiet --depth 1 --branch "$v" \
  https://github.com/charmbracelet/bubbletea ".charm-repos/bubbletea@$v"

v=$(go list -m -f '{{.Version}}' charm.land/lipgloss/v2)
[ -d ".charm-repos/lipgloss@$v" ] || git -c advice.detachedHead=false clone \
  --quiet --depth 1 --branch "$v" \
  https://github.com/charmbracelet/lipgloss ".charm-repos/lipgloss@$v"
```

Both tags are annotated, so git prints a benign `refs/tags/vX is not a commit!`
warning on a shallow clone. Ignore it — the clone is correct.

`.charm-repos/` is gitignored and sits at the **repo root**; all paths below
are relative to it. The version in the directory name is what keeps the clone
honest — after a `go.mod` bump the path simply won't exist yet, so re-running
the above re-clones at the new tag.

## How to use

1. Find the closest example/tutorial in the maps below — rows are bare
   directory names.
2. Resolve it to a path: Bubble Tea rows live under
   `bubbletea@<version>/examples/` (or `tutorials/` where noted), Lip Gloss
   rows under `lipgloss@<version>/examples/`.
3. `Read` its `main.go` (and `README.md` when present) to copy the v2 idiom,
   then adapt.

Full index: `bubbletea@<version>/examples/README.md`.

## Reading library source

Examples show idiom; for exact API shape (signatures, struct fields, return
types) read the **pinned source in the module cache** — it always matches
`go.mod`, unlike anything vendored:

- `go doc charm.land/bubbles/v2/list` — quick signature lookup.
- `go list -m -f '{{.Dir}}' charm.land/bubbles/v2` — prints the cache dir;
  `Read` files under it for full source.

## Learn the fundamentals first

- `tutorials/basics/` — Model/Init/Update/View, key handling, the core loop.
- `tutorials/commands/` — `tea.Cmd`, async work, custom messages.

## Bubble Tea — by need

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

## Lip Gloss — by need

| Need | Example dir |
|---|---|
| Styles, borders, layout | `layout` |
| Styled tables | `table/` (e.g. `table/languages`) |
| Styled lists | `list/` (e.g. `list/simple`) |
| Trees | `tree/` (e.g. `tree/simple`) |
| Color / adaptive color | `color`, `compat` |
| Gradients / blending | `blending/`, `brightness` |
| Free-form drawing | `canvas` |
