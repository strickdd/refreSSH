package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDispatcher(t *testing.T) {
	prefix := "ctrl+a"
	d := NewDispatcher(prefix)

	// Test prefix enters command mode
	action, consumed := d.Handle(tea.KeyMsg{Type: tea.KeyCtrlA})
	if !consumed {
		t.Error("Prefix should be consumed")
	}
	if action != ActionNone {
		t.Errorf("Expected ActionNone, got %v", action)
	}
	if !d.InCommand() {
		t.Error("Should be in command mode")
	}

	// Test mapped key in command mode
	action, consumed = d.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !consumed {
		t.Error("Mapped key should be consumed")
	}
	if action != ActionNextTab {
		t.Errorf("Expected ActionNextTab, got %v", action)
	}
	if d.InCommand() {
		t.Error("Should not be in command mode anymore")
	}

	// Test prefix twice sends prefix
	d.Handle(tea.KeyMsg{Type: tea.KeyCtrlA})
	action, consumed = d.Handle(tea.KeyMsg{Type: tea.KeyCtrlA})
	if !consumed {
		t.Error("Prefix should be consumed")
	}
	if action != ActionSendPrefix {
		t.Errorf("Expected ActionSendPrefix, got %v", action)
	}

	// Test non-prefix key passes through when not in command mode
	action, consumed = d.Handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if consumed {
		t.Error("Non-prefix key should not be consumed when not in command mode")
	}
	if action != ActionNone {
		t.Errorf("Expected ActionNone, got %v", action)
	}
}
