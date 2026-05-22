# Architecture Decision Records

One record per pinned decision. `DECISIONS.md` (repo root) is the short,
always-current summary; these ADRs are the longer-form context + consequences
behind each line of it. When a decision changes, supersede the ADR (add a new
one that references the old) rather than silently editing history.

Format per file: title, status, date, context, decision, consequences.

| # | Decision |
|---|----------|
| [0001](0001-git-wrapper.md) | All git access goes through a tested wrapper |
| [0002](0002-fixed-typed-command-set.md) | A fixed, typed command set — view *and* mutate |
| [0003](0003-discovery-immediate-subdirs.md) | Discovery scans immediate subdirectories of one root |
| [0004](0004-fzf-style-interaction.md) | fzf-style: always-focused filter, non-printable actions |
| [0005](0005-commands-act-on-filtered-set.md) | Commands act on the filtered set; no multi-select |
| [0006](0006-no-confirmation.md) | Commands fire immediately, without confirmation |
| [0007](0007-checkout-via-switch-guess.md) | checkout uses a transient prompt + `git switch` guess |
| [0008](0008-inline-results-and-refresh.md) | Inline results, per-row auto-refresh, manual `ctrl+r` |
| [0009](0009-compact-row-plus-drill-in.md) | Compact main row plus a drill-in detail view |
| [0010](0010-concurrency-message-driven.md) | Concurrency: one `tea.Cmd` per repo, message-driven |
| [0011](0011-thin-vertical-slices.md) | Build in thin vertical slices; reviews on request |
| [0012](0012-tui-testing.md) | TUI tested end-to-end, with model-level tests for in-place changes |
| [0013](0013-logging-to-file.md) | Logging to `~/gbx.log`; the TUI owns stdout |
