Require micasa

Output images/pr-documents.gif

Set Shell bash
Set FontFamily "JetBrains Mono"
Set FontSize 24
Set Width 2200
Set Height 1200
Set Padding 20
Set Theme "Dracula"
Set CursorBlink false
Set TypingSpeed 0

Env NO_COLOR ""
Env TERM "xterm-256color"
Env COLORTERM "truecolor"
Env COLORFGBG "15;0"
Env PS1 ""

Hide
Type "exec micasa --demo"
Enter
Sleep 5s

# Dismiss dashboard and move to Documents tab.
Type "D"
Sleep 1s
Show
Sleep 0.5s
Tab 5
Sleep 2s

# Enter edit mode and open add-document form.
Type "i"
Sleep 0.8s
Type "a"
Sleep 2s

# Move to the file picker field and open picker mode.
Tab
Sleep 0.8s
Enter
Sleep 2s

# Type to filter and then exit picker/form.
Type "doc"
Sleep 1.8s
Escape
Sleep 1s
Escape
Sleep 0.8s
Escape
Sleep 0.8s

# Quit app.
Type "q"
Sleep 1s
