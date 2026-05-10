package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/strickdd/refressh/internal/daemon"
)

func TestSessionEndpoints(t *testing.T) {
	d := daemon.New(nil)
	handler := NewHandler(d)

	cmd := "sh"
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
	}

	// 1. Test POST /sessions
	createReq := map[string]interface{}{
		"id":      "test-api-session",
		"command": cmd,
		"args":    []string{},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/sessions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. Test GET /sessions
	req = httptest.NewRequest("GET", "/sessions", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var sessions []*daemon.Session
	if err := json.NewDecoder(w.Body).Decode(&sessions); err != nil {
		t.Fatalf("Failed to decode sessions: %v", err)
	}

	found := false
	for _, s := range sessions {
		if s.ID == "test-api-session" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created session not found in list")
	}

	// 3. Test DELETE /sessions/{id}
	req = httptest.NewRequest("DELETE", "/sessions/test-api-session", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify it's gone
	req = httptest.NewRequest("GET", "/sessions", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if err := json.NewDecoder(w.Body).Decode(&sessions); err != nil {
		t.Fatalf("Failed to decode sessions: %v", err)
	}
	for _, s := range sessions {
		if s.ID == "test-api-session" {
			t.Error("Session still exists after DELETE")
		}
	}
}
