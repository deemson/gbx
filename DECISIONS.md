# gbx — Decisions

The pinned decisions for gbx. Read this before building; update it when a decision changes.

## What gbx is

A TUI to view the state of many git repositories at once **and** run a fixed set
of git commands across them.

## Product

- **Shows state and mutates.** Not read-only.
- **Fixed command set**, not arbitrary passthrough. Each command is a typed
  method on the `internal/git` wrapper, so output is structured and testable.
  v1 commands: `pull`, `checkout`.
- **Discovery:** scan the *immediate* subdirectories of one root dir; each that
  is a git repo becomes a row, non-repos are ignored. Root comes from a CLI arg,
  default current working dir. No config file yet; no recursive scan.

## Interaction (fzf-style)

- A **filter text input is always focused** from startup. Printable keys go to
  the filter; **every action is a non-printable binding** (ctrl/alt/enter/esc).
- Filter **fuzzy-matches** repo names.
- A command acts on **the repos currently matching the filter**. **No marking /
  multi-select.** Clearing the filter targets all repos.
- Commands are invoked via **dedicated `ctrl+` bindings**, avoiding
  terminal-reserved combos (`ctrl+c/z/s/q`, and `ctrl+m/i/h/[` which are
  enter/tab/backspace/esc).
- **Commands fire immediately — no confirmation step.**
- `checkout` needs an argument: its binding opens a **transient `branch:`
  prompt** (enter runs / esc cancels). It uses `git switch <branch>` with the
  **default guess** (auto-creates a local tracking branch from a same-named
  remote). Failures are detected by **attempt-and-read** — run the command, map
  non-zero exit + stderr to a typed error → the row shows `✗`.

## Display

One table, one row per repo:

`name | branch (marker if no upstream) | ↑ahead ↓behind | worktree state | [result]`

- Worktree state = `clean` or a changed-path count, with a conflict flag.
- Raw lines `+X/-Y` are **not** in the main row — they live in the drill-in view.
- The result cell shows a spinner while a command runs → `✓`/`✗`.

## Results & refresh

- Command results are shown **inline in the table** — no separate log pane.
- A row's git state **auto-refreshes** after a command completes on it.
- **`ctrl+r`** triggers a manual refresh. **No periodic background refresh** in v1.
- Quit: `esc` / `ctrl+c`.

## Concurrency

Repos are opened and queried **concurrently** — one `tea.Cmd` per repo, results
arrive as messages that update individual rows (the pattern proven in the earlier
`tui2` work).

## Process

- **Requirements-light, then code in thin vertical slices.** Each slice is one
  visible capability end-to-end, tested with `gitest` + the `testProgram`
  harness, and committed before the next.
- **Reviews on request**, not automated: `code-review` per slice, and
  `security-review` because mutating commands (`pull`/`checkout`) can lose work.

## Deferred / out of scope

- Arbitrary command passthrough.
- `commit`, `push`, branch creation (`-b`).
- Config-file repo lists; recursive discovery.
- The `just simulate` recipe / golden snapshots — abandoned. Ignore the orphaned
  recipe in the `justfile`.

## Slice roadmap

0. Scaffold + docs.
1. Discovery + filterable list.
2. Read-state columns.
3. `pull`.
4. `checkout`.
5. Drill-in detail + help overlay + keymap doc + per-decision ADRs.
