# TUI Framework Research

## Candidates

### 1. Bubble Tea (Charm.sh)
- **Model:** Elm Architecture (Model-View-Update).
- **Pros:**
    - Highly composable and modular.
    - Excellent for complex state management (like "Command Mode").
    - Very active ecosystem (Lip Gloss for styling, Bubbles for components).
    - Functional approach makes it highly testable and revisable.
- **Cons:**
    - Steep learning curve for those unfamiliar with Elm.
    - Can become verbose for simple layouts.

### 2. Tview
- **Model:** Widget-based (traditional OOP-like).
- **Pros:**
    - Rich set of built-in high-level widgets (Tables, Forms, Lists).
    - Easier for traditional layouts (Flex, Grid).
    - Familiar for those coming from other TUI frameworks.
- **Cons:**
    - Less flexible for "experimental" UI patterns like Naomi's "Command Mode".
    - State management can be trickier in complex apps compared to Bubble Tea.

## Recommendation: Bubble Tea

For refreSSH, **Bubble Tea** is the recommended framework. 

### Rationale:
1. **Command Mode Integration:** The Elm architecture natively supports the state transitions required for a "Command Mode" (similar to Vim).
2. **Revisability:** The separation of model, view, and update logic allows Naomi to iterate on the UI without breaking core functionality.
3. **Component Ecosystem:** Charm's `bubbles` already has many of the primitives we need, and styling with `lipgloss` is much more intuitive for modern TUIs.
4. **Daemon Integration:** Bubble Tea's message-passing system will integrate well with the `broadcaster` and PTY logic in `internal/daemon`.

## Prototype Implementation
A basic prototype is available in `internal/tui/prototype.go` and can be run via `cmd/tui-prototype/main.go`.

### Features:
- **Tab Bar:** Dynamic session switching via `Tab`.
- **Command Mode:** Toggle with `:` and return with `Esc`/`Enter`.
- **Styling:** Uses `lipgloss` for consistent aesthetics.
