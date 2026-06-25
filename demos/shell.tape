# gbx shell-action clip — regenerate with `just demo shell`.
# Films against the throwaway repo tree built by `just gen-demo-fixture`; the
# justfile runs VHS from inside that directory with gbx on PATH and overrides
# the output path, so the Output line below only matters if you run it by hand.
# Assumes the default actions menu (1 lazygit, 2 shell); `2` fires the shell.
Output assets/shell.gif
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

# Narrow to the one dirty repo (web-frontend) and commit the filter.
Sleep 800ms
Ctrl+F
Sleep 600ms
Type "web"
Sleep 1.2s
Enter
Sleep 800ms

# Open its actions menu and launch a shell in the repo's directory.
Enter
Sleep 1.5s
Type "2"
Sleep 1.5s

# Inspect the dirty tree, then drop back into gbx.
Type "git status"
Enter
Sleep 2.5s
Type "exit"
Enter
Sleep 2s
