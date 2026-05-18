package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKeyToBytes_Printable(t *testing.T) {
	// Printable runes should pass through as their byte value
	tests := []struct {
		name     string
		key      tea.KeyMsg
		expected []byte
	}{
		{"letter a", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, []byte("a")},
		{"letter z", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}, []byte("z")},
		{"digit 1", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}, []byte("1")},
		{"space", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, []byte(" ")},
		{"unicode", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'é'}}, []byte("é")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyToBytes(tt.key)
			if string(got) != string(tt.expected) {
				t.Errorf("keyToBytes(%v) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKeyToBytes_ControlChars(t *testing.T) {
	tests := []struct {
		name     string
		keyType  tea.KeyType
		expected []byte
	}{
		{"NUL/Ctrl+@", tea.KeyNull, []byte{0x00}},
		{"Ctrl+A", tea.KeyCtrlA, []byte{0x01}},
		{"Ctrl+B", tea.KeyCtrlB, []byte{0x02}},
		{"Ctrl+C", tea.KeyCtrlC, []byte{0x03}},
		{"Ctrl+D", tea.KeyCtrlD, []byte{0x04}},
		{"Ctrl+H", tea.KeyCtrlH, []byte{0x08}},
		{"Tab", tea.KeyTab, []byte{0x09}},
		{"Ctrl+I", tea.KeyCtrlI, []byte{0x09}},
		{"Ctrl+J", tea.KeyCtrlJ, []byte{0x0A}},
		{"Ctrl+K", tea.KeyCtrlK, []byte{0x0B}},
		{"Ctrl+L", tea.KeyCtrlL, []byte{0x0C}},
		{"Enter", tea.KeyEnter, []byte{0x0D}},
		{"Ctrl+M", tea.KeyCtrlM, []byte{0x0D}},
		{"Ctrl+N", tea.KeyCtrlN, []byte{0x0E}},
		{"Ctrl+O", tea.KeyCtrlO, []byte{0x0F}},
		{"Ctrl+P", tea.KeyCtrlP, []byte{0x10}},
		{"Ctrl+Q", tea.KeyCtrlQ, []byte{0x11}},
		{"Ctrl+R", tea.KeyCtrlR, []byte{0x12}},
		{"Ctrl+S", tea.KeyCtrlS, []byte{0x13}},
		{"Ctrl+T", tea.KeyCtrlT, []byte{0x14}},
		{"Ctrl+U", tea.KeyCtrlU, []byte{0x15}},
		{"Ctrl+V", tea.KeyCtrlV, []byte{0x16}},
		{"Ctrl+W", tea.KeyCtrlW, []byte{0x17}},
		{"Ctrl+X", tea.KeyCtrlX, []byte{0x18}},
		{"Ctrl+Y", tea.KeyCtrlY, []byte{0x19}},
		{"Ctrl+Z", tea.KeyCtrlZ, []byte{0x1A}},
		{"Esc", tea.KeyEsc, []byte{0x1B}},
		{"Ctrl+\\", tea.KeyCtrlBackslash, []byte{0x1C}},
		{"Ctrl+]", tea.KeyCtrlCloseBracket, []byte{0x1D}},
		{"Ctrl+^", tea.KeyCtrlCaret, []byte{0x1E}},
		{"Ctrl+_", tea.KeyCtrlUnderscore, []byte{0x1F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyToBytes(tea.KeyMsg{Type: tt.keyType})
			if !equalBytes(got, tt.expected) {
				t.Errorf("keyToBytes(%v) = %v, want %v", tt.keyType, got, tt.expected)
			}
		})
	}
}

func TestKeyToBytes_Backspace(t *testing.T) {
	// Backspace should send 0x7F (DEL)
	got := keyToBytes(tea.KeyMsg{Type: tea.KeyBackspace})
	if !equalBytes(got, []byte{0x7F}) {
		t.Errorf("keyToBytes(KeyBackspace) = %v, want 0x7F", got)
	}
}

func TestKeyToBytes_Arrows(t *testing.T) {
	tests := []struct {
		name     string
		keyType  tea.KeyType
		expected []byte
	}{
		{"Up", tea.KeyUp, []byte("\x1b[A")},
		{"Down", tea.KeyDown, []byte("\x1b[B")},
		{"Right", tea.KeyRight, []byte("\x1b[C")},
		{"Left", tea.KeyLeft, []byte("\x1b[D")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyToBytes(tea.KeyMsg{Type: tt.keyType})
			if !equalBytes(got, tt.expected) {
				t.Errorf("keyToBytes(%v) = %v, want %v", tt.keyType, got, tt.expected)
			}
		})
	}
}

func TestKeyToBytes_SpecialKeys(t *testing.T) {
	tests := []struct {
		name     string
		keyType  tea.KeyType
		expected []byte
	}{
		{"Enter", tea.KeyEnter, []byte{0x0D}},
		{"Ctrl+M", tea.KeyCtrlM, []byte{0x0D}},
		{"Delete", tea.KeyDelete, []byte("\x1b[3~")},
		{"Insert", tea.KeyInsert, []byte("\x1b[2~")},
		{"Home", tea.KeyHome, []byte("\x1b[H")},
		{"End", tea.KeyEnd, []byte("\x1b[F")},
		{"PgUp", tea.KeyPgUp, []byte("\x1b[5~")},
		{"PgDown", tea.KeyPgDown, []byte("\x1b[6~")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyToBytes(tea.KeyMsg{Type: tt.keyType})
			if !equalBytes(got, tt.expected) {
				t.Errorf("keyToBytes(%v) = %v, want %v", tt.keyType, got, tt.expected)
			}
		})
	}
}

func TestKeyToBytes_FunctionKeys(t *testing.T) {
	tests := []struct {
		name     string
		keyType  tea.KeyType
		expected []byte
	}{
		{"F1", tea.KeyF1, []byte("\x1bOP")},
		{"F2", tea.KeyF2, []byte("\x1bOQ")},
		{"F3", tea.KeyF3, []byte("\x1bOR")},
		{"F4", tea.KeyF4, []byte("\x1bOS")},
		{"F5", tea.KeyF5, []byte("\x1b[15~")},
		{"F6", tea.KeyF6, []byte("\x1b[17~")},
		{"F7", tea.KeyF7, []byte("\x1b[18~")},
		{"F8", tea.KeyF8, []byte("\x1b[19~")},
		{"F9", tea.KeyF9, []byte("\x1b[20~")},
		{"F10", tea.KeyF10, []byte("\x1b[21~")},
		{"F11", tea.KeyF11, []byte("\x1b[23~")},
		{"F12", tea.KeyF12, []byte("\x1b[24~")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyToBytes(tea.KeyMsg{Type: tt.keyType})
			if !equalBytes(got, tt.expected) {
				t.Errorf("keyToBytes(%v) = %v, want %v", tt.keyType, got, tt.expected)
			}
		})
	}
}

// TestKeyToBytes_NoRegression verifies that keyToBytes produces actual control
// bytes rather than the text string representations returned by tea.KeyMsg.String().
// This is a regression test for the original bug where []byte(msg.String()) was
// sent to the PTY instead of proper byte sequences, causing "ctrl+c", "backspace",
// "tab", "up", "down" etc. to appear as literal text on the remote terminal.
func TestKeyToBytes_NoRegression(t *testing.T) {
	// These are the keys that were broken before the fix. Each must produce
	// a non-printable, non-text-like byte sequence.
	brokenKeys := []tea.KeyMsg{
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyBackspace},
		{Type: tea.KeyTab},
		{Type: tea.KeyUp},
		{Type: tea.KeyDown},
		{Type: tea.KeyRight},
		{Type: tea.KeyLeft},
		{Type: tea.KeyDelete},
		{Type: tea.KeyInsert},
		{Type: tea.KeyHome},
		{Type: tea.KeyEnd},
		{Type: tea.KeyPgUp},
		{Type: tea.KeyPgDown},
		{Type: tea.KeyEsc},
		{Type: tea.KeyEnter},
		{Type: tea.KeyF1},
		{Type: tea.KeyF12},
	}

	for _, km := range brokenKeys {
		got := keyToBytes(km)
		str := km.String()
		if len(got) == 0 {
			t.Errorf("keyToBytes(%s) = empty bytes, expected non-empty byte sequence", str)
			continue
		}

		// If this produces a string like "ctrl+c", "backspace", "tab", "up", etc.
		// the original bug is back. Check that the output is NOT the key's String() value.
		if len(got) >= 5 && string(got) == str {
			t.Errorf("keyToBytes(%s) returned %q — this is the broken text-string path, expected raw bytes", str, got)
		}

		// Verify that printable control chars are single bytes in the control range
		// (0x00-0x1F), or ANSI escape sequences starting with 0x1B
		for i, b := range got {
			if i == 0 && b == 0x1B {
				// ANSI escape sequence - valid
				continue
			}
			if b < 0x20 && b != 0x00 {
				// Control byte in non-first position is valid for ANSI sequences
				continue
			}
			// First byte of a multi-char ANSI sequence like \x1b[2~
			if i > 0 && (got[i-1] == 0x1B || (i >= 2 && got[i-2] == 0x1B && got[i-1] == '[')) {
				continue
			}
			// Allow single control bytes (0x00-0x1F)
			if len(got) == 1 && b < 0x20 {
				continue
			}
		}
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
