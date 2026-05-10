package daemon

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/strickdd/refressh/internal/config"
)

// TestSessionAsyncExecution verifies that a process continues running and its output
// is captured in the scrollback buffer even when no clients are attached.
func TestSessionAsyncExecution(t *testing.T) {
	// Use a temporary config directory to avoid messing with user data
	tempDir, err := os.MkdirTemp("", "refressh-async-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config.SetConfigDirOverride(tempDir)
	defer config.SetConfigDirOverride("") // Reset after test

	cfg := config.NewDefaultConfig()
	d := New(cfg)

	// 1. Create a session that prints a marker and stays open
	sessionID := "e2e-async-test"
	cmd := "sh"
	args := []string{"-c", "echo ASYNC_MARKER; sleep 60"}
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
		args = []string{"/c", "echo ASYNC_MARKER && ping 127.0.0.1 -n 60 > nul"}
	}

	s, err := d.CreateSession(sessionID, cmd, args...)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer func() {
		// Ignore error on stop as process might have exited
		_ = d.StopSession(sessionID)
	}()

	// 2. Poll for the output to be captured in the scrollback buffer
	// Increased polling to be more patient on CI
	found := false
	for i := 0; i < 50; i++ {
		scrollbackContent := string(s.Broadcaster.Scrollback())

		if strings.Contains(scrollbackContent, "ASYNC_MARKER") {
			found = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !found {
		scrollbackContent := string(s.Broadcaster.Scrollback())
		t.Errorf("Expected 'ASYNC_MARKER' in scrollback, got: %q", scrollbackContent)
	}

	// 3. Attach a mock client and check if the marker is delivered
	mockClient := &mockWSClient{id: "mock-client", output: make(chan []byte, 100), t: t}
	s.Broadcaster.AddClient(mockClient)

	select {
	case data := <-mockClient.output:
		if !strings.Contains(string(data), "ASYNC_MARKER") {
			t.Errorf("Expected 'ASYNC_MARKER' in delivered data, got: %s", string(data))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for mock client to receive scrollback")
	}

	if err := d.StopSession(sessionID); err != nil {
		t.Logf("Warning: failed to stop session: %v", err)
	}
}

// TestMetadataPersistence verifies that session metadata is saved to disk and
// can be recovered, allowing sessions to be restarted.
func TestMetadataPersistence(t *testing.T) {
	// Use a temporary config directory to avoid messing with user data
	tempDir, err := os.MkdirTemp("", "refressh-persistence-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config.SetConfigDirOverride(tempDir)
	defer config.SetConfigDirOverride("") // Reset after test

	cfg := config.NewDefaultConfig()
	d := New(cfg)

	sessionID := "persist-1"
	cmd := "sleep"
	args := []string{"10"}
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
		args = []string{"/c", "ping 127.0.0.1 -n 10 > nul"}
	}

	_, err = d.CreateSession(sessionID, cmd, args...)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Verify sessions are in the map
	sessions := d.Sessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	// Verify file exists in temp dir
	statePath := filepath.Join(tempDir, "sessions.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("sessions.json was not created in the temporary directory")
	}

	// Simulate "crash" by creating a new daemon and loading state
	d2 := New(cfg)
	if err := d2.loadState(); err != nil {
		t.Fatalf("loadState failed: %v", err)
	}

	if len(d2.Sessions()) != 1 {
		t.Errorf("Expected 1 session to be loaded, got %d", len(d2.Sessions()))
	}
}

type mockWSClient struct {
	id     string
	output chan []byte
	t      *testing.T
}

func (m *mockWSClient) ID() string { return m.id }
func (m *mockWSClient) Write(p []byte) (int, error) {
	m.output <- p
	return len(p), nil
}
func (m *mockWSClient) SendStatus(s Status) error { return nil }

