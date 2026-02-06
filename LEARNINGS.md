# Learnings

- Do NOT `cd` into the workspace directory before running commands -- you're already in it.
- The color palette must be colorblind-safe (Wong palette) and must use
  `lipgloss.AdaptiveColor{Light, Dark}` for every color so it auto-adjusts to
  the terminal's background. When adding or changing styles, always provide both
  Light and Dark variants. See `styles.go` for the existing palette and the
  comment block documenting chromatic/neutral roles.
