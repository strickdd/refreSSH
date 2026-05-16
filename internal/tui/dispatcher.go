package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Action represents a TUI action that can be triggered by a key binding in command mode.
type Action int

const (
	// ActionNone represents no action.
	ActionNone Action = iota
	// ActionQuit quits the application.
	ActionQuit
	// ActionNextTab moves to the next tab.
	ActionNextTab
	// ActionPrevTab moves to the previous tab.
	ActionPrevTab
	// ActionNewTab creates a new tab.
	ActionNewTab
	// ActionCloseTab closes the current tab.
	ActionCloseTab
	// ActionSendPrefix sends the prefix key to the PTY.
	ActionSendPrefix
	// ActionDetach detaches from the current session.
	ActionDetach
	// ActionScrollbackSearch opens scrollback in a pager.
	ActionScrollbackSearch
)

// Dispatcher handles key interceptions and command mode logic.
type Dispatcher struct {
	prefix      string
	inCommand   bool
	keyMappings map[string]Action
}

// NewDispatcher creates a new dispatcher with a configurable prefix.
func NewDispatcher(prefix string) *Dispatcher {
	return &Dispatcher{
		prefix:    prefix,
		inCommand: false,
		keyMappings: map[string]Action{
			"n":      ActionNextTab,
			"p":      ActionPrevTab,
			"c":      ActionNewTab,
			"x":      ActionCloseTab,
			"q":      ActionQuit,
			"d":      ActionDetach,
			"D":      ActionDetach,
			"s":      ActionScrollbackSearch,
			"S":      ActionScrollbackSearch,
		},
	}
}

// Handle handles a key message and returns an action and whether the message was consumed.
func (d *Dispatcher) Handle(msg tea.KeyMsg) (Action, bool) {
	keyStr := msg.String()

	if !d.inCommand {
		if keyStr == d.prefix {
			d.inCommand = true
			return ActionNone, true
		}
		return ActionNone, false
	}

	// We are in command mode
	d.inCommand = false // Always exit command mode after next key
	if keyStr == d.prefix {
		return ActionSendPrefix, true
	}
	if action, ok := d.keyMappings[keyStr]; ok {
		return action, true
	}

	// If key is not mapped, we still consume it (it just exits command mode)
	return ActionNone, true
}

// InCommand returns true if the dispatcher is currently waiting for a command key.
func (d *Dispatcher) InCommand() bool {
	return d.inCommand
}
