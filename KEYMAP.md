# gbx keymap

gbx is fzf-style: a filter input is always focused, so **printable keys filter**
and **every action is a non-printable binding**. A command acts on the repos
currently matching the filter (clear the filter to target all).

| Key | Action |
|-----|--------|
| _(type)_ | filter repos (fuzzy) |
| `↑` / `↓` (or `ctrl+k` / `ctrl+j`) | move the cursor |
| `enter` | drill into the repo under the cursor |
| `ctrl+p` | pull the filtered repos |
| `ctrl+o` | checkout a branch on the filtered repos (opens a `branch:` prompt) |
| `ctrl+r` | refresh status of the filtered repos |
| `ctrl+g` | toggle the help overlay |
| `esc` | back (from detail / help / prompt), or quit from the list |
| `ctrl+c` | quit |

Bindings deliberately avoid terminal-reserved combos (`ctrl+c/z/s/q`, and
`ctrl+m/i/h/[` which are enter/tab/backspace/esc).

The in-app `ctrl+g` overlay renders the same list from `keyBindings` in
`internal/tui/help.go` — that slice is the single source of truth; keep this
table in sync with it.
