# gbx

A TUI to view the state of many git repositories at once **and** run a fixed set
of git commands across them.

## Scope

- **A fixed command set, not free-form.** List mode is the default — letter keys
  dispatch typed `Repo` methods directly on the filtered repos: `r` refresh,
  `f` fetch, `p` pull, `c` Switch Branch (arg prompt with branch autocomplete
  drawn from the union across the visible repos; `tab`/`shift+tab` cycle), `b`
  New Branch (arg prompt; same suggestion source as `c` for reference, Tab
  cycles — picking an existing name fails on Enter and the typed error surfaces
  on the row). `?` toggles the help overlay; `ctrl+f` opens the filter prompt
  (Enter commits the draft to the active filter; ESC clears the draft, or —
  when already empty — reverts and closes; ctrl+f while open reverts). `q` (or
  `ctrl+c` anywhere) quits. Each row has a **2-wide left gutter** before the
  name: a dim spinner while the row is busy (reading status/diff/branches **or**
  running a command), a red `✗` once it settles with an error, blank otherwise —
  **success is silent** (no `✓`). A failed row also gets a **one-liner** that is
  the typed error (`err.Error()`); the command error wins over a load error. An
  explicit `r` refresh clears a settled error (gutter + one-liner) before
  re-reading; a command's own follow-up refresh keeps it. The error is also
  logged to `~/gbx.log`. There is **no output pane** — the typed errors are the
  surface.
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
  field (name+branch / name / branch). The app has a **three-row header at the
  top** that's always visible: row 1 is `Filter: <value>` (dim `none` when
  empty), row 2 is the field chips `<C-1> name + branch · <C-2> name · <C-3>
  branch` — each chip a dim `<C-N>` key prefix plus a label, the active chip's
  label bold + accent — and row 3 is a full-width dim `─` rule. The right corner
  is static dim chrome on rows 1–2: `gbx <version>` over `PID: <pid>`, shown in
  every mode (`version` defaults to `dev`; release builds set it via
  `WithVersion`/ldflags). `ctrl+f` opens the filter
  prompt — row 1 becomes the live-editable draft and live-narrows the visible
  rows (Enter commits to `m.filter`). `c` opens the Switch Branch prompt
  (`Switch Branch:` on row 1, branch suggestions on row 2, dim `(no matches)`
  when the draft narrows them to empty); `b` opens the New Branch prompt
  (`New Branch:` on row 1, same suggestion source as reference). `c`/`b` lack
  the retrigger-close — their letters are typeable in refs/branch names.
  `ctrl+1/2/3` are unbound in `c`/`b` prompts (field stays sticky). `?`
  toggles the help overlay (alt screen). There is no bottom bar. The binding
  slices in `internal/tui/help.go` are the single source of truth for what `?`
  documents.
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
  surface (the row's gutter `✗` + error one-liner). Tests discard logs (see `TestMain`).

## Build / run / test

- `go build` → `./gbx`; run `./gbx [root-dir]` (default: cwd). The header
  version shows `dev` for a plain build; release builds stamp it via
  `go build -ldflags "-X main.version=v1.2.3"`.
- `go test ./...`

## Workflow

- Code + test each change, then commit before the next.
- The command set is non-destructive (`checkout` refuses to overwrite local
  changes; `pull` is `--ff-only`) and runs with no confirmation step — run
  `code-review` / `security-review` on request.
