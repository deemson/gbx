# gbx

A TUI to view the state of many git repositories at once **and** run a fixed set
of git commands across them.

## Scope

- **A fixed command set, not free-form.** List mode is the default — letter keys
  dispatch typed `Repo` methods directly on the filtered repos: `r` refresh,
  `f` fetch, `p` pull, `c` checkout (opens an arg prompt with branch autocomplete
  drawn from the union across the visible repos; `tab`/`shift+tab` cycle), `b`
  checkout -b (arg prompt, no autocomplete). `F1` toggles the help overlay; `F4`
  opens the filter prompt (Enter commits the draft to the active filter; ESC
  clears the draft, or — when already empty — reverts and closes; F4 while open
  reverts). `q` (or `ctrl+c` anywhere) quits. Per-repo result is a `⟳/✓/✗`
  glyph **plus a one-liner** that, on failure, is the typed error
  (`err.Error()`); success shows nothing. The error is also logged to
  `~/gbx.log`. There is **no output pane** — the typed errors are the surface.
- **Discovery:** scan the *immediate* subdirectories of one root dir (CLI arg,
  default cwd); each that is a git repo becomes a row. No recursion, no config
  file.
- A command acts on **the repos currently matching the filter** — no marking /
  multi-select, and **no confirmation step**. Clearing the filter targets all.
- **Out of scope:** config-file repo lists; recursive discovery.

## Layout

- `internal/git` — the **tested git wrapper**. This is the foundation. Every git
  action is a typed method on `Repo`: structured reads (`Status`,
  `DiffNumStatHead`, `Branches`) and the command set (`Checkout`,
  `CheckoutBranch`, `Fetch`, `Pull`), all mapping errors by attempt-and-read.
  **Do not shell out to `git` anywhere else** — there is no generic runner.
  - `internal/git/exec` — raw `git` process runner.
  - `internal/git/gitest` — test helpers that build real repos (`Init`, `Clone`,
    `WriteFileAdd`, `Commit`, `Push`, `Pull`, `Fetch`, …). Use these for tests
    across the whole codebase, not just the `git` package.
- `internal/tui` — the Bubble Tea v2 app (`charm.land/bubbletea/v2`, `bubbles/v2`,
  `lipgloss/v2`).
- `main.go` — wires logging (→ `~/gbx.log`) and runs the TUI with the root dir.

## Conventions

- **Extend the git wrapper, don't bypass it.** A new git action = a new typed
  method on `Repo`, with errors mapped by **attempt-and-read** (inspect exit code
  + stderr → typed error), as in `open.go` / `diff_numstat.go` / `repo.go`. Never
  shell out elsewhere.
- **The TUI is htop-style:** list mode is the default — letter keys dispatch
  git actions directly on the filtered repos and `ctrl+1/2/3` toggle the filter
  field (name+branch / name / branch). `F4` opens a transient filter prompt at
  the bottom row; while it's open, the draft live-narrows the visible rows
  (Enter commits to `m.filter`). `c` and `b` open argument prompts with the same
  state machine, minus the retrigger-close — `c`/`b` are typeable in
  refs/branch names. `F1` toggles the help overlay (alt screen). The bottom bar
  shows the committed filter on the left (or empty when none) and `F1 Help`
  pinned to the right corner; while a prompt is open, the prompt input replaces
  the left half. The binding slices in `internal/tui/help.go` are the single
  source of truth for what `F1` documents.
- **Test the TUI end-to-end** with the `testProgram` harness (`internal/tui`,
  `testhelper_test.go`): it drives a real `tea.Program`, inject keys with
  `send`/`sendKey`, assert rendered output with `waitForContent`. Build fixtures
  with `gitest`. **Caveat:** the alt-screen renderer does differential,
  cursor-positioned updates, so `waitForContent` only reliably sees *fresh/
  appended* text — an in-place change (e.g. `↓1`→`↓0`) is not a contiguous
  substring. Assert state *transitions* with renderer-free model-level tests
  (drive `model.Update` directly, inspect state), as in `model_test.go`.
- **Logging:** zerolog → `~/gbx.log` (the TUI owns stdout). Each command's
  outcome (the typed error, or success) is logged here, in addition to its in-app
  surface (the row glyph + error one-liner). Tests discard logs (see `TestMain`).

## Build / run / test

- `go build` → `./gbx`; run `./gbx [root-dir]` (default: cwd).
- `go test ./...`

## Workflow

- Code + test each change, then commit before the next.
- The command set is non-destructive (`checkout` refuses to overwrite local
  changes; `pull` is `--ff-only`) and runs with no confirmation step — run
  `code-review` / `security-review` on request.
