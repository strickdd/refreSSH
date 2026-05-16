package tui

import (
	"testing"
)

func TestRingBuffer_Write(t *testing.T) {
	rb := newRingBuffer(100)

	n, err := rb.Write([]byte("hello"))
	if n != 5 {
		t.Errorf("expected n=5, got %d", n)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got := string(rb.Bytes())
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestRingBuffer_MaxSize(t *testing.T) {
	rb := newRingBuffer(10)

	_, _ = rb.Write([]byte("0123456789abcdef"))

	got := string(rb.Bytes())
	expected := "6789abcdef"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRingBuffer_Concurrent(t *testing.T) {
	rb := newRingBuffer(10000)
	done := make(chan struct{})

	go func() {
		for i := 0; i < 1000; i++ {
			rb.Write([]byte("x"))
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = rb.Bytes()
		}
		done <- struct{}{}
	}()

	<-done
	<-done

	n := len(rb.Bytes())
	if n != 1000 {
		t.Errorf("expected 1000 bytes, got %d", n)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"hello", []string{"hello"}},
		{"hello\nworld", []string{"hello", "world"}},
		{"line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"trailing\n", []string{"trailing"}},
	}

	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("splitLines(%q): expected %v, got %v", tt.input, tt.expected, got)
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitLines(%q)[%d]: expected %q, got %q", tt.input, i, tt.expected[i], got[i])
			}
		}
	}
}

func TestReorderTabsMRU(t *testing.T) {
	m := Model{
		tabs: []*Tab{
			{SessionID: "a"},
			{SessionID: "b"},
			{SessionID: "c"},
		},
		mruOrder:       make(map[string]int),
		activeTabIndex: 0,
	}

	m.reorderTabsMRU()

	if len(m.tabs) != 3 {
		t.Fatalf("expected 3 tabs, got %d", len(m.tabs))
	}
	if m.tabs[0].SessionID != "b" {
		t.Errorf("expected active tab 'b' at index 0, got %q", m.tabs[0].SessionID)
	}
	if m.tabs[2].SessionID != "a" {
		t.Errorf("expected moved tab 'a' at index 2, got %q", m.tabs[2].SessionID)
	}
	if m.activeTabIndex != 2 {
		t.Errorf("expected activeTabIndex=2, got %d", m.activeTabIndex)
	}
}

func TestReorderTabsMRU_NoTabs(t *testing.T) {
	m := Model{
		tabs:           []*Tab{},
		mruOrder:       make(map[string]int),
		activeTabIndex: -1,
	}

	m.reorderTabsMRU()

	if len(m.tabs) != 0 {
		t.Errorf("expected 0 tabs, got %d", len(m.tabs))
	}
}
