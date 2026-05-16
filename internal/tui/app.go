// Package tui implements the terminal user interface for refreSSH.
package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/strickdd/refressh/internal/config"
)

// WSControlMessage is sent from client to daemon as a JSON text message.
type WSControlMessage struct {
	Action string `json:"action"`
}

// WSStatus is received from daemon as a JSON text message.
type WSStatus struct {
	IsPrimary bool `json:"is_primary"`
}

// Tab represents a single terminal tab with its own WebSocket connection.
type Tab struct {
	SessionID    string
	Conn         *websocket.Conn
	IsPrimary    bool
	Connected    bool
	Disconnected bool
	Title        string
	buffer       string
	mu           sync.Mutex
}

// ringBuffer is a simple thread-safe byte ring buffer for scrollback history.
type ringBuffer struct {
	buf []byte
	mu  sync.Mutex
	max int
}

func newRingBuffer(maxSize int) *ringBuffer {
	return &ringBuffer{
		buf: make([]byte, 0, maxSize),
		max: maxSize,
	}
}

func (r *ringBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf = append(r.buf, p...)
	if len(r.buf) > r.max {
		r.buf = r.buf[len(r.buf)-r.max:]
	}
	return len(p), nil
}

func (r *ringBuffer) Bytes() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]byte, len(r.buf))
	copy(res, r.buf)
	return res
}

// Model represents the state of the TUI application.
type Model struct {
	dispatcher        *Dispatcher
	tabs              []*Tab
	mruOrder          map[string]int
	activeTabIndex    int
	width             int
	height            int
	connected         bool
	quit              bool
	err               error
	pagerActive       bool
	pagerBuffer       []byte
	pagerMsg          string
	apiURL            string
	token             string
	scrollbackPager   *ringBuffer
	viewport          viewport.Model
	availableSessions []SessionSummary
}

// SessionSummary is a lightweight session representation for listing.
type SessionSummary struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Running bool   `json:"running"`
}

const (
	maxScrollback = 1024 * 1024
	connWriteWait = 10 * time.Second
)

//go:generate go run golang.org/x/tools/go/analysis/passes/embedcmd/cmd/generate -tag embed

func defaultShell() string {
	switch runtime.GOOS {
	case "windows":
		return "pwsh"
	default:
		return "bash"
	}
}

// InitialModel initializes the TUI model.
func InitialModel() Model {
	cfg, err := config.Load()
	if err != nil {
		return Model{err: fmt.Errorf("failed to load config: %w", err)}
	}

	token, err := config.GetAPIToken()
	if err != nil {
		return Model{err: fmt.Errorf("failed to load API token: %w", err)}
	}

	port := cfg.Port
	if port == 0 {
		port = 8080
	}

	return Model{
		dispatcher:        NewDispatcher("ctrl+b"),
		tabs:              make([]*Tab, 0),
		mruOrder:          make(map[string]int),
		apiURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
		token:             token,
		scrollbackPager:   newRingBuffer(maxScrollback),
		availableSessions: make([]SessionSummary, 0),
		viewport:          viewport.New(80, 24),
	}
}

func (m *Model) fetchSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		m.fetchSessions()
		return nil
	}
}

func (m *Model) fetchSessions() {
	u := fmt.Sprintf("%s/sessions", m.apiURL)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var sessions []SessionSummary
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return
	}
	m.availableSessions = sessions
}

func (m *Model) createNewTabCmd(sessionID *string) tea.Cmd {
	return func() tea.Msg {
		m.createSessionAndConnect(sessionID)
		return nil
	}
}

func (m *Model) createSessionAndConnect(sessionID *string) {
	var sid string

	if sessionID != nil && *sessionID != "" {
		sid = *sessionID
	} else {
		if len(m.availableSessions) > 0 {
			sid = m.availableSessions[0].ID
		} else {
			sid = fmt.Sprintf("session-%d", time.Now().UnixNano())
			go func() {
				reqBody, _ := json.Marshal(map[string]interface{}{
					"id":      sid,
					"command": defaultShell(),
				})
				req, _ := http.NewRequest("POST", fmt.Sprintf("%s/sessions", m.apiURL), bytes.NewBuffer(reqBody))
				req.Header.Set("Authorization", "Bearer "+m.token)
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return
				}
		defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusCreated {
					m.fetchSessions()
				}
			}()
		}
	}

	params := url.Values{}
	params.Add("id", sid)
	wsURL := url.URL{
		Scheme:   "ws",
		Host:     fmt.Sprintf("127.0.0.1:%d", m.getPort()),
		Path:     "/attach",
		RawQuery: params.Encode(),
	}

	header := http.Header{}
	header.Add("Authorization", "Bearer "+m.token)

	tab := &Tab{
		SessionID: sid,
		Title:     sid,
	}

	m.tabs = append(m.tabs, tab)
	m.mruOrder[sid] = len(m.mruOrder)
	m.activeTabIndex = len(m.tabs) - 1
	m.connected = true

	go func() {
		dialer := websocket.Dialer{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
		}
		conn, _, err := dialer.Dial(wsURL.String(), header)
		if err != nil {
			tab.Connected = false
			tab.Disconnected = true
			return
		}
		tab.Conn = conn
		tab.Connected = true
		m.readLoop(tab)
	}()
}

func (m *Model) readLoop(tab *Tab) {
	defer func() {
		if tab.Conn != nil {
			_ = tab.Conn.Close()
		}
		tab.Connected = false
		tab.Disconnected = true
	}()

	for {
		msgType, message, err := tab.Conn.ReadMessage()
		if err != nil {
			return
		}

		if msgType == websocket.TextMessage {
			var status WSStatus
			if err := json.Unmarshal(message, &status); err == nil {
				tab.IsPrimary = status.IsPrimary
			}
			continue
		}

		// Binary message: PTY output
		tab.mu.Lock()
		tab.buffer += string(message)
		tab.mu.Unlock()

		_, _ = m.scrollbackPager.Write(message)
	}
}

func (m *Model) getPort() int {
	apiURL := m.apiURL
	if len(apiURL) < 7 {
		return 8080
	}
	portStr := apiURL[7:]
	colonIdx := bytes.LastIndex([]byte(portStr), []byte(":"))
	if colonIdx == -1 {
		return 8080
	}
	var port int
	_, _ = fmt.Sscanf(portStr[colonIdx+1:], "%d", &port)
	return port
}

func (m Model) activeTab() *Tab {
	if m.activeTabIndex < 0 || m.activeTabIndex >= len(m.tabs) {
		return nil
	}
	return m.tabs[m.activeTabIndex]
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.fetchSessionsCmd()
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quit {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.pagerActive {
			switch msg.String() {
			case "q", "ctrl+c":
				m.pagerActive = false
				m.pagerBuffer = nil
				m.pagerMsg = ""
			case " ":
				m.viewport.ViewDown()
			case "home":
				m.viewport.GotoTop()
			case "end":
				m.viewport.GotoBottom()
			}
			return m, nil
		}

		action, consumed := m.dispatcher.Handle(msg)
		if consumed {
			return m.handleAction(action)
		}

		if tab := m.activeTab(); tab != nil && tab.Connected && !tab.Disconnected {
			if tab.IsPrimary {
				return m, m.sendToTab(tab, []byte(msg.String()))
			}
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "ctrl+space" {
			return m, m.requestPrimary()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentHeight := m.height - 6
		if contentHeight < 0 {
			contentHeight = 0
		}
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = contentHeight

		// Send resize to active tab
		if tab := m.activeTab(); tab != nil && tab.Connected && !tab.Disconnected {
			return m, m.sendResize(tab, contentHeight, msg.Width-2)
		}
	}

	return m, nil
}

func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionQuit:
		m.quit = true
		return m, tea.Quit

	case ActionNextTab:
		if len(m.tabs) > 1 {
			m.activeTabIndex = (m.activeTabIndex + 1) % len(m.tabs)
			m.reorderTabsMRU()
		}

	case ActionPrevTab:
		if len(m.tabs) > 1 {
			m.activeTabIndex = (m.activeTabIndex - 1 + len(m.tabs)) % len(m.tabs)
			m.reorderTabsMRU()
		}

	case ActionNewTab:
		if len(m.availableSessions) > 0 {
			sid := m.availableSessions[0].ID
			return m, m.createNewTabCmd(&sid)
		}
		return m, m.createNewTabCmd(nil)

	case ActionCloseTab:
		if len(m.tabs) > 1 {
			tab := m.tabs[m.activeTabIndex]
			if tab.Conn != nil {
				_ = tab.Conn.Close()
			}
			go func() {
				u := fmt.Sprintf("%s/sessions/%s", m.apiURL, tab.SessionID)
				req, _ := http.NewRequest("DELETE", u, nil)
				req.Header.Set("Authorization", "Bearer "+m.token)
				resp, err := http.DefaultClient.Do(req)
				if err == nil && resp != nil {
					_ = resp.Body.Close()
				}
			}()

			m.tabs = append(m.tabs[:m.activeTabIndex], m.tabs[m.activeTabIndex+1:]...)
			delete(m.mruOrder, tab.SessionID)
			if m.activeTabIndex >= len(m.tabs) {
				m.activeTabIndex = len(m.tabs) - 1
			}
		}

	case ActionSendPrefix:
		if tab := m.activeTab(); tab != nil && tab.Connected && !tab.Disconnected && tab.IsPrimary {
			return m, m.sendToTab(tab, []byte{0x02})
		}

	case ActionDetach:
		if tab := m.activeTab(); tab != nil && tab.Conn != nil {
			_ = tab.Conn.Close()
		}
		if len(m.tabs) > 1 {
			m.tabs = append(m.tabs[:m.activeTabIndex], m.tabs[m.activeTabIndex+1:]...)
			if m.activeTabIndex >= len(m.tabs) {
				m.activeTabIndex = len(m.tabs) - 1
			}
		}
		m.pagerMsg = "Detached from session"
		return m, tea.Batch(m.fetchSessionsCmd(), m.clearPagerMsg())

	case ActionScrollbackSearch:
		m.pagerActive = true
		m.pagerBuffer = m.scrollbackPager.Bytes()
		m.pagerMsg = "Scrollback (q to quit, space to page, home/end to jump)"
	}

	return m, nil
}

func (m Model) sendToTab(tab *Tab, data []byte) tea.Cmd {
	return func() tea.Msg {
		if tab.Conn == nil || tab.Disconnected {
			return nil
		}
		_ = tab.Conn.SetWriteDeadline(time.Now().Add(connWriteWait))
		err := tab.Conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			return WSError{Tab: tab, Err: err}
		}
		return nil
	}
}

func (m Model) requestPrimary() tea.Cmd {
	tab := m.activeTab()
	if tab == nil || tab.Conn == nil {
		return nil
	}
	req := WSControlMessage{Action: "request_primary"}
	data, _ := json.Marshal(req)
	return func() tea.Msg {
		_ = tab.Conn.SetWriteDeadline(time.Now().Add(connWriteWait))
		err := tab.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return WSError{Tab: tab, Err: err}
		}
		return nil
	}
}

func (m Model) clearPagerMsg() tea.Cmd {
	return func() tea.Msg {
		m.pagerMsg = ""
		return nil
	}
}

func (m Model) sendResize(tab *Tab, rows, cols int) tea.Cmd {
	return func() tea.Msg {
		if tab.Conn == nil || tab.Disconnected {
			return nil
		}
		resizeMsg := map[string]int{"rows": rows, "cols": cols}
		data, _ := json.Marshal(resizeMsg)
		_ = tab.Conn.SetWriteDeadline(time.Now().Add(connWriteWait))
		err := tab.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return WSError{Tab: tab, Err: err}
		}
		return nil
	}
}

func (m *Model) reorderTabsMRU() {
	if m.activeTabIndex < 0 || m.activeTabIndex >= len(m.tabs) {
		return
	}
	tab := m.tabs[m.activeTabIndex]
	if tab == nil {
		return
	}
	m.tabs = append(m.tabs[:m.activeTabIndex], m.tabs[m.activeTabIndex+1:]...)
	m.tabs = append(m.tabs, tab)
	m.activeTabIndex = len(m.tabs) - 1
}


func (m Model) renderPager() string {
	doc := ""

	if m.pagerMsg != "" {
		doc += m.pagerMsg + "\n"
	}

	content := string(m.pagerBuffer)
	content = stripAnsi(content)

	height := m.height - 2
	if m.pagerMsg != "" {
		height--
	}
	if height <= 0 {
		height = 1
	}

	if len(content) > height*80 {
		lines := splitLines(content)
		if len(lines) > height {
			content = lipgloss.NewStyle().Height(height).Render(
				lipgloss.JoinVertical(lipgloss.Left, lines[:height]...),
			)
		} else {
			content = lipgloss.NewStyle().Height(height).Render(
				lipgloss.JoinVertical(lipgloss.Left, lines...),
			)
		}
	}

	doc += content

	return doc
}

func splitLines(s string) []string {
	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// WSError represents a WebSocket error message.
type WSError struct {
	Tab *Tab
	Err error
}

// Err returns the initialization error if the model failed to load.
func (m Model) Err() error {
	return m.err
}
