// Package tui implements the terminal user interface for refreSSH.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the state of the TUI application.
type Model struct {
	dispatcher     *Dispatcher
	tabs           []string
	activeTabIndex int
	terminal       string
	width          int
	height         int
}

// InitialModel initializes the TUI model.
func InitialModel() Model {
	return Model{
		dispatcher:     NewDispatcher("ctrl+a"),
		tabs:           []string{"Session 1"},
		activeTabIndex: 0,
		terminal:       "Welcome to refreSSH\nPrefix is Ctrl+A\n- Ctrl+A, c: New Tab\n- Ctrl+A, n: Next Tab\n- Ctrl+A, p: Previous Tab\n- Ctrl+A, x: Close Tab\n- Ctrl+A, q: Quit",
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles TUI messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		action, consumed := m.dispatcher.Handle(msg)
		if consumed {
			return m.handleAction(action)
		}
		// If not consumed, it would normally go to the active terminal/pty
		// For now, we just handle some global keys if not in command mode
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionQuit:
		return m, tea.Quit
	case ActionNextTab:
		if len(m.tabs) > 0 {
			m.activeTabIndex = (m.activeTabIndex + 1) % len(m.tabs)
		}
	case ActionPrevTab:
		if len(m.tabs) > 0 {
			m.activeTabIndex = (m.activeTabIndex - 1 + len(m.tabs)) % len(m.tabs)
		}
	case ActionNewTab:
		m.tabs = append(m.tabs, "New Session")
		m.activeTabIndex = len(m.tabs) - 1
	case ActionCloseTab:
		if len(m.tabs) > 1 {
			m.tabs = append(m.tabs[:m.activeTabIndex], m.tabs[m.activeTabIndex+1:]...)
			if m.activeTabIndex >= len(m.tabs) {
				m.activeTabIndex = len(m.tabs) - 1
			}
		}
	case ActionSendPrefix:
		// In a real app, this would send the prefix to the PTY
		m.terminal += "\n[Prefix Sent]"
	}
	return m, nil
}
