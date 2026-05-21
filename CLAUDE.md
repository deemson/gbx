# gbx

A TUI for viewing and operating on many git repositories at once.
**Read `DECISIONS.md`** for the product/architecture decisions and the slice roadmap.

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
  filter; every action is a non-printable binding.
- **Test the TUI end-to-end** with the `testProgram` harness (`internal/tui`,
  `testhelper_test.go`): it drives a real `tea.Program` over piped I/O; assert
  rendered output with `waitForContent`. Build fixtures with `gitest`.
- **Logging:** zerolog → `~/gbx.log` (the TUI owns stdout). Tests discard logs
  (see `TestMain`).

## Build / run / test

- `go build` → `./gbx`; run `./gbx [root-dir]` (default: cwd).
- `go test ./...`
- The `just simulate` recipe is dead — ignore it.

## Workflow

- Build in thin vertical slices (see `DECISIONS.md` roadmap); each slice is
  coded + tested, then committed before the next.
- Run `code-review` / `security-review` on request.
