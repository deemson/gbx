# gbx open multiple-directories clip — regenerate with `just demo open`.
# Films the nested services/tools fleets built by the shared demo fixture.
Output assets/open.gif
Require gbx

Set Shell "bash"
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 18
Set Width 1200
Set Height 560
Set Padding 20
Set Theme "Catppuccin Mocha"

# Beat 1 — CLI order and exact argument spelling become gutter-free headings.
Hide
Type "gbx open ./services ./tools"
Enter
Sleep 3s
Show
Sleep 1.5s

# Beat 2 — filtering changes repo rows, not headings; tools has no API match.
Ctrl+F
Sleep 400ms
Type "api"
Sleep 2s
