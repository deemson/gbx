# gbx

A TUI to view the state of many git repos at once and run a fixed git command
set across the ones currently matching the filter.

Everything below is what the code can't tell you. If reading the code answers
it, it doesn't belong here — don't add it back.

- **Scope is deliberately small and fixed — not free-form.** Discovery is a flat
  scan of the cwd; the command set is non-destructive and runs with no
  confirmation step. Don't add recursion, config-driven repo lists, or a second
  path to run git — every git action is a typed method on the one wrapper.
- **Before editing TUI code, invoke the `charm-tui` skill.** The stack is the v2
  charm libraries; your recalled API is the wrong major version — imports,
  signatures, and message types all differ.
- **The alt-screen renderer redraws differentially**, so the test harness only
  reliably sees fresh/appended text. Assert state *transitions* at the model
  level, never in-place character changes.
