# gbx

A TUI to view the state of many git repositories at once **and** run arbitrary
git commands across them.

## Scope

- **Arbitrary git command execution.** `tab` toggles the input line into command
  mode; the typed line (shell-style quoting, leading `git` optional) runs against
  the filtered repos via the generic `Repo.Run`. Per-repo result is a `⟳/✓/✗`
  glyph (by exit code); the full stdout/stderr goes to `~/gbx.log`, **not** the
  UI. Structured data shown in the table (status, `+/-` line changes) still comes
  from typed wrapper methods.
- **Discovery:** scan the *immediate* subdirectories of one root dir (CLI arg,
  default cwd); each that is a git repo becomes a row. No recursion, no config
  file.
- A command acts on **the repos currently matching the filter** — no marking /
  multi-select, and **no confirmation step**. Clearing the filter targets all.
- **Out of scope:** in-app command output (it goes to the log); config-file repo
  lists; recursive discovery.

## Layout

- `internal/git` — the **tested git wrapper**. This is the foundation. Structured
  reads are typed methods on `Repo` (`Status`, `DiffNumStatHead`); arbitrary
  execution goes through the generic `Repo.Run(ctx, args...) (exec.Result, error)`.
  **Do not shell out to `git` anywhere else.**
  - `internal/git/exec` — raw `git` process runner.
  - `internal/git/gitest` — test helpers that build real repos (`Init`, `Clone`,
    `WriteFileAdd`, `Commit`, `Push`, `Pull`, `Fetch`, …). Use these for tests
    across the whole codebase, not just the `git` package.
- `internal/tui` — the Bubble Tea v2 app (`charm.land/bubbletea/v2`, `bubbles/v2`,
  `lipgloss/v2`).
- `main.go` — wires logging (→ `~/gbx.log`) and runs the TUI with the root dir.

## Conventions

- **Extend the git wrapper, don't bypass it.** A new *structured* read = a new
  typed method on `Repo`, with errors mapped by **attempt-and-read** (inspect exit
  code + stderr → typed error), as in `open.go` / `diff_numstat.go`. Everything
  else runs through the generic `Repo.Run` — still never shell out elsewhere.
- **The TUI is fzf-style:** in the list a filter input is always focused —
  printable keys filter, every action is a non-printable binding. `tab` toggles
  command mode, where the same line edits a git command (printable keys edit it)
  run on the filtered repos. The `ctrl+g` help overlay renders from `keyBindings`
  in `internal/tui/help.go` — that slice is the single source of truth.
- **Test the TUI end-to-end** with the `testProgram` harness (`internal/tui`,
  `testhelper_test.go`): it drives a real `tea.Program`, inject keys with
  `send`/`sendKey`, assert rendered output with `waitForContent`. Build fixtures
  with `gitest`. **Caveat:** the alt-screen renderer does differential,
  cursor-positioned updates, so `waitForContent` only reliably sees *fresh/
  appended* text — an in-place change (e.g. `↓1`→`↓0`) is not a contiguous
  substring. Assert state *transitions* with renderer-free model-level tests
  (drive `model.Update` directly, inspect state), as in `model_test.go`.
- **Logging:** zerolog → `~/gbx.log` (the TUI owns stdout). Command output
  (exit/stdout/stderr) is logged here — it has no in-app surface. Tests discard
  logs (see `TestMain`).

## Build / run / test

- `go build` → `./gbx`; run `./gbx [root-dir]` (default: cwd).
- `go test ./...`

## Workflow

- Code + test each change, then commit before the next.
- Arbitrary commands can lose work (e.g. `reset --hard`, `clean -fd`) and run with
  no confirmation — run `code-review` / `security-review` on request.
