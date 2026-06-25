# gbx overview clip — regenerate with `just demo overview`.
# Films against the throwaway repo tree built by `just gen-demo-fixture`; the
# justfile runs VHS from inside that directory with gbx on PATH and overrides
# the output path, so the Output line below only matters if you run it by hand.
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

# The fleet at a glance: branch, ahead/behind, and dirty markers on every repo.
Sleep 3.5s
