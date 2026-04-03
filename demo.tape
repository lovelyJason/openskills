# OpenSkills Demo — VHS Tape File
# Usage: vhs demo.tape
#
# 生成 GIF 和 MP4:
#   vhs demo.tape
#
# 发布到 vhs.charm.sh 获取链接:
#   vhs publish screenshots/osk-demo.gif

Output screenshots/osk-demo.gif
Output screenshots/osk-demo.mp4

Set Shell zsh
Set FontSize 15
Set Width 1200
Set Height 900
Set Theme "Dracula"
Set WindowBar Colorful
Set WindowBarSize 40
Set Padding 24
Set TypingSpeed 50ms
Set CursorBlink false

Sleep 300ms

# Show system status first
Type "osk status"
Sleep 100ms
Enter
Sleep 1200ms

# Then show the full list
Type "osk list"
Sleep 100ms
Enter
Sleep 2500ms

Sleep 500ms

# Interactive skill install — select claude editor
Type "osk skill install dev-debug"
Sleep 200ms
Enter
Sleep 1500ms

# MultiSelect: "Select target editors:" — order: codex, claude, cursor
# Move down to "claude"
Down
Sleep 300ms
# Space to toggle select "claude"
Space
Sleep 300ms
# Enter to confirm selection
Enter
Sleep 1000ms

# Install mode prompt: symlink (default) / native
# Just press Enter to accept symlink
Enter
Sleep 3s

Sleep 1s

# Enter Claude REPL to show the slash command is available
Type "claude"
Sleep 200ms
Enter
Sleep 3s

# Type the slash command (no Enter — just demonstrate it exists)
Type "/dev-debug"
Sleep 2s
