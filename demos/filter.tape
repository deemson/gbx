# gbx filter clip — regenerate with `just demo filter`.
# Films against the throwaway repo tree built by `just gen-demo-fixture`; the
# justfile runs VHS from inside that directory with gbx on PATH and overrides
# the output path, so the Output line below only matters if you run it by hand.
Output assets/filter.gif
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

# Open the fzf-style filter and narrow to the repos on main.
Sleep 800ms
Ctrl+F
Sleep 800ms
Type "main"
Sleep 2s

# Invert the term with `!`: everything NOT on main.
Backspace 4
Type "!main"
Sleep 2s

# Commit the filter and settle.
Enter
Sleep 1.2s
