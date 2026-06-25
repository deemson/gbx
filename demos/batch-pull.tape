# gbx batch-pull clip — regenerate with `just demo batch-pull`.
# Films against the throwaway repo tree built by `just gen-demo-fixture`; the
# justfile runs VHS from inside that directory with gbx on PATH and overrides
# the output path, so the Output line below only matters if you run it by hand.
Output assets/batch-pull.gif
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

# Filter to the repos on main, then commit.
Sleep 800ms
Ctrl+F
Sleep 600ms
Type "main"
Sleep 1.5s
Enter
Sleep 1s

# Pull the whole filtered set: notifications fast-forwards, the remoteless
# repo reports an error (red ✗), the rest hold steady.
Type "p"
Sleep 3.5s
