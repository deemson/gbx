# gbx overview clip — regenerate with `just demo overview`.
# Films against the throwaway repo tree built by `just gen-demo-fixture`; the
# justfile runs VHS from inside that directory with gbx on PATH and overrides
# the output path, so the Output line below only matters if you run it by hand.
# A fast montage of the main features: filter, search, and a batch pull.
Output assets/overview.gif
Require gbx

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

# Launch; hide discovery + status load so only the settled fleet is filmed.
Hide
Type "gbx"
Enter
Sleep 3s
Show

# Beat 1 — the fleet at a glance: branch, ahead/behind, and dirty markers.
Sleep 2.5s

# Beat 2 — filter narrows the list. `!main` (a DSL negation) drops everything on
# main, leaving just the two off-main repos; then clear back to the full fleet.
Ctrl+F
Sleep 600ms
Type "!main"
Sleep 1.8s
Escape
Sleep 400ms
Escape
Sleep 800ms

# Beat 3 — search keeps every repo visible and jumps the cursor between matches.
# `main` lights the four repos on main and dims the rest; ↓ walks the cursor,
# leaping past the dimmed rows. Enter keeps the landing spot.
Ctrl+S
Sleep 600ms
Type "main"
Sleep 1.2s
Down
Sleep 800ms
Down
Sleep 800ms
Enter
Sleep 800ms

# Beat 4 — pull the whole fleet: the behind repo fast-forwards (green ✓), the
# remoteless one reports an error (red ✗), the rest hold steady.
Type "p"
Sleep 3.5s
