# ═══════════════════════════════════════════════════════════════════
# Paymail Inspector Demo
#
# Renders examples/demo.gif (the README demo). Run from the repository root,
# with network access, since every scene talks to live paymail providers:
#
#   vhs examples/demo.tape
#
# The hidden setup builds the CLI into ./bin (gitignored) and adds the
# `paymail` symlink, exactly like the README installer does.
# ═══════════════════════════════════════════════════════════════════

Output examples/demo.gif

Require go

# ═══════════════════════════════════════════════════════════════════
# VISUAL SETTINGS
# ═══════════════════════════════════════════════════════════════════

Set Shell "zsh"
Set FontSize 13
Set Width 1560
Set Height 800
Set Padding 20
Set BorderRadius 10
Set WindowBar Colorful
Set TypingSpeed 45ms
Set LineHeight 1.2

# ═══════════════════════════════════════════════════════════════════
# THEME - Paymail gold on midnight
# ═══════════════════════════════════════════════════════════════════

Set Theme {
  "name": "PaymailMidnight",
  "black": "#0b1020",
  "red": "#fb7185",
  "green": "#34d399",
  "yellow": "#fbbf24",
  "blue": "#60a5fa",
  "magenta": "#f472b6",
  "cyan": "#22d3ee",
  "white": "#e2e8f0",
  "brightBlack": "#64748b",
  "brightRed": "#fda4af",
  "brightGreen": "#6ee7b7",
  "brightYellow": "#fcd34d",
  "brightBlue": "#93c5fd",
  "brightMagenta": "#f9a8d4",
  "brightCyan": "#67e8f9",
  "brightWhite": "#f8fafc",
  "background": "#0b1020",
  "foreground": "#e2e8f0",
  "selection": "#fbbf2455",
  "cursor": "#fbbf24"
}

# ═══════════════════════════════════════════════════════════════════
# SETUP (hidden)
# ═══════════════════════════════════════════════════════════════════

Hide
Type "go build -o bin/paymail-inspector ./cmd/paymail-inspector && ln -sf paymail-inspector bin/paymail"
Enter
Sleep 15s
Type "export PATH=$PWD/bin:$PATH NO_UPDATE_CHECK=1 PS1='%F{yellow}❯%f ' && paymail --flush-cache >/dev/null && clear"
Enter
Sleep 1s
Show

# ═══════════════════════════════════════════════════════════════════
# DEMO SCRIPT
# ═══════════════════════════════════════════════════════════════════

# Scene 1: Version
Sleep 500ms
Type "paymail --version"
Sleep 300ms
Enter
Sleep 1.5s

Hide
Type "clear"
Enter
Show

# Scene 2: Capability discovery (SRV record -> host -> capabilities, with BRFC names)
Type "paymail capabilities handcash.io -t"
Sleep 300ms
Enter
Sleep 5s

Hide
Type "clear"
Enter
Show

# Scene 3: Resolve an address (with request tracing)
Type "paymail resolve mrz@relayx.io"
Sleep 300ms
Enter
Sleep 6s

Hide
Type "clear"
Enter
Show

# Scene 4: Whois across every public provider, in parallel
Type "paymail whois mrz"
Sleep 300ms
Enter
Sleep 6s

Hide
Type "clear"
Enter
Show

# Scene 5: Validate a domain against the bsvalias specification
Type "paymail validate relayx.io -t --skip-dnssec"
Sleep 300ms
Enter
Sleep 6s

Hide
Type "clear"
Enter
Show

# Scene 6: Generate a BRFC ID
Type 'paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1'
Sleep 300ms
Enter
Sleep 3s
