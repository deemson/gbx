# gbx

A TUI to view the state of many git repositories at once **and** run a fixed set
of git commands across them.

## Scope

- **Fixed, typed command set** — not arbitrary passthrough. Each command is a
  typed method on the `internal/git` wrapper, so its output is structured and
  testable. v1 commands: `pull`, `checkout`.
- **Discovery:** scan the *immediate* subdirectories of one root dir (CLI arg,
  default cwd); each that is a git repo becomes a row. No recursion, no config
  file.
- A command acts on **the repos currently matching the filter** — no marking /
  multi-select, and **no confirmation step**. Clearing the filter targets all.
- **Out of scope:** arbitrary command passthrough; `commit` / `push` / branch
  creation (`-b`); config-file repo lists; recursive discovery.

## Layout

- `internal/git` — the **tested git wrapper**. This is the foundation. Add every
  new git operation here as a typed method on `Repo` (like `Status`,
  `DiffNumStatHead`). **Do not shell out to `git` anywhere else.**
  - `internal/git/exec` — raw `git` process runner.
  - `internal/git/gitest` — test helpers that build real repos (`Init`, `Clone`,
    `WriteFileAdd`, `Commit`, `Push`, `Pull`, `Fetch`, …). Use these for tests
    across the whole codebase, not just the `git` package.
- `internal/tui` — the Bubble Tea v2 app (`charm.land/bubbletea/v2`, `bubbles/v2`,
  `lipgloss/v2`).
- `main.go` — wires logging (→ `~/gbx.log`) and runs the TUI with the root dir.

## Conventions

- **Extend the git wrapper, don't bypass it.** A new git command = a new typed
  method on `Repo`, with errors mapped by **attempt-and-read** (inspect exit code
  + stderr → typed error), as in `open.go` / `diff_numstat.go`.
- **The TUI is fzf-style:** a filter input is always focused. Printable keys
  filter; every action is a non-printable binding. The `ctrl+g` help overlay
  renders from `keyBindings` in `internal/tui/help.go` — that slice is the single
  source of truth for the bindings.
- **Test the TUI end-to-end** with the `testProgram` harness (`internal/tui`,
  `testhelper_test.go`): it drives a real `tea.Program`, inject keys with
  `send`/`sendKey`, assert rendered output with `waitForContent`. Build fixtures
  with `gitest`. **Caveat:** the alt-screen renderer does differential,
  cursor-positioned updates, so `waitForContent` only reliably sees *fresh/
  appended* text — an in-place change (e.g. `↓1`→`↓0`) is not a contiguous
  substring. Assert state *transitions* with renderer-free model-level tests
  (drive `model.Update` directly, inspect state), as in `model_test.go`.
- **Logging:** zerolog → `~/gbx.log` (the TUI owns stdout). Tests discard logs
  (see `TestMain`).

## Build / run / test

- `go build` → `./gbx`; run `./gbx [root-dir]` (default: cwd).
- `go test ./...`

## Workflow

- Code + test each change, then commit before the next.
- Mutating commands (`pull` / `checkout`) can lose work — run `code-review` /
  `security-review` on request.
