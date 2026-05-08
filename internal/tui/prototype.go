// Package tui provides the terminal user interface for refreSSH.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	id          string
	commandMode bool
}

// InitialModel returns the initial TUI model state.
func InitialModel() model {
	return model{
		id: "default",
	}
}

// Init initializes the TUI model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles TUI messages and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+a":
			m.commandMode = !m.commandMode
			return m, nil
		}
	}
	return m, nil
}

// View renders the TUI.
func (m model) View() string {
	s := "refreSSH - Persistent Terminal\n\n"
	if m.commandMode {
		s += "COMMAND MODE ACTIVE\n"
	} else {
		s += "Terminal Area (Normal Mode)\n"
	}
	s += "\nPress Ctrl+A to toggle command mode, Q to quit.\n"
	return s
}
