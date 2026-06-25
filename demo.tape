# gbx demo — regenerate with `just demo`.
# Films against the throwaway repo tree built by `just gen-demo-fixture`; the
# justfile runs VHS from inside that directory with gbx on PATH.
Output demo.gif
Require gbx

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Dracula"

# Launch and let discovery + status load settle. Hidden so the shell prompt and
# launch keystrokes aren't part of the recording.
Hide
Type "gbx"
Enter
Sleep 3s
Show

# The fleet at a glance: branches, ahead/behind, dirty markers across every repo.
Sleep 2.5s

# Narrow the fleet with the fzf-style filter.
Ctrl+F
Sleep 800ms
Type "notif"
Sleep 1.5s
Enter
Sleep 1.5s

# Run a batch git action across the filtered set: pull fast-forwards it.
Type "p"
Sleep 2.5s

# Open the actions menu for the cursored repo (lazygit / shell).
Enter
Sleep 2s
Escape
Sleep 1s

# Quit.
Type "q"
Sleep 1.5s
