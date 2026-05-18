# TUI Framework

## Selected Framework: Bubble Tea (Charm.sh)

**Model:** Elm Architecture (Model-View-Update).

- **Highly composable and modular** — excellent for complex state management like Command Mode.
- **Active ecosystem** — Lip Gloss for styling, Bubbles for reusable components.
- **Functional approach** — makes testing and iteration straightforward.
- **Message-passing** — integrates well with the daemon's broadcaster and PTY logic.

## Current Implementation

The TUI is implemented in `internal/tui/` using Bubble Tea with:
- **Tab Bar:** Dynamic session switching.
- **Command Mode:** Toggle with `:` and return with `Esc`/`Enter`.
- **Styling:** Uses `lipgloss` for consistent aesthetics.
