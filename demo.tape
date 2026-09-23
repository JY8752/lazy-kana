# Recreate the README demo with: vhs demo.tape
Output assets/demo.gif

Require go

Set Shell "zsh"
Set FontSize 18
Set Width 960
Set Height 600
Set Padding 20
Set Framerate 10
Set TypingSpeed 80ms
Set Theme "Catppuccin Mocha"

# Build and launch without recording the setup commands.
Hide
Type "go build -o /tmp/lazy-kana-vhs-demo ."
Enter
Sleep 3s
Type "clear"
Enter
Type "/tmp/lazy-kana-vhs-demo"
Enter
Sleep 1s
Show

# Menu and oversized dot-art kana.
Sleep 1s
Type "d"
Sleep 1s
Type "1"
Sleep 2s
Space
Sleep 1s
Space
Sleep 1s
Escape
Sleep 1s

# Add several names to the roster.
Type "2"
Sleep 1s
Type@120ms "たろう"
Enter
Sleep 700ms
Type@120ms "はな"
Enter
Sleep 700ms
Type@120ms "ゆうき"
Enter
Sleep 1s
Escape
Sleep 1s

# Finish with the animated hiragana train.
Type "5"
Sleep 4s
Space
Sleep 3s
Ctrl+C

Hide
Type "rm -f /tmp/lazy-kana-vhs-demo"
Enter
