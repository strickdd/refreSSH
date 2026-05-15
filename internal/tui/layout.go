package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/ansi"
)

var (
	tabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), true, true, false, true)

	activeTabStyle = tabStyle.
			BorderForeground(lipgloss.Color("62")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	commandModeStyle = statusStyle.
				Background(lipgloss.Color("160"))

	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62"))
)

// View renders the TUI to a string for display.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	doc := strings.Builder{}

	// Show error if present
	if m.err != nil {
		doc.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		return doc.String()
	}

	// Pager overlay
	if m.pagerActive {
		doc.WriteString(m.renderPager())
		return doc.String()
	}

	// Tabs row
	var renderedTabs []string
	for i, tab := range m.tabs {
		title := tab.SessionID
		style := tabStyle
		if i == m.activeTabIndex {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(fmt.Sprintf("%d:%s", i+1, title)))
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(tabsRow + "\n")

	// Terminal area
	contentHeight := m.height - 6 // Tabs + Status bar + Borders
	if contentHeight < 0 {
		contentHeight = 0
	}

	tab := m.activeTab()

	// Get terminal content
	var terminalContent string
	if tab == nil {
		terminalContent = "No active session"
	} else if !tab.Connected || tab.Disconnected {
		terminalContent = "Disconnected\n\nReconnect: Ctrl+B, c - New Tab"
	} else {
		tab.mu.Lock()
		terminalContent = tab.buffer
		tab.mu.Unlock()
	}

	// Strip ANSI escapes for lipgloss rendering
	terminalContent = stripAnsi(terminalContent)

	// Set viewport content for scrolling support
	m.viewport.SetContent(terminalContent)

	rendered := windowStyle.
		Width(m.width - 2).
		Height(contentHeight).
		Render(m.viewport.View())
	doc.WriteString(rendered + "\n")

	// Status bar
	modeStr := " NORMAL "
	style := statusStyle
	if m.dispatcher.InCommand() {
		modeStr = " COMMAND "
		style = commandModeStyle
	}

	status := style.Render(modeStr)
	if m.dispatcher.InCommand() {
		status += " [d]etach [s]crollback [Ctrl+B] literal "
	} else {
		status += fmt.Sprintf(" %s - Tab %d/%d", m.dispatcher.prefix, m.activeTabIndex+1, len(m.tabs))
		if tab != nil && !tab.Disconnected && !tab.IsPrimary {
			status += " | Ctrl+Space: Request Control"
		}
	}

	// Truncate status to fit width
	if len(status) > m.width-2 {
		status = status[:m.width-2]
	}

	doc.WriteString(status)

	return doc.String()
}

// stripAnsi removes ANSI escape sequences from the given string.
// This wraps the muesli/ansi.Strip function for use in rendering.
func stripAnsi(s string) string {
	return ansiStripper(s)
}

// ansiStripper is a var for testability.
var ansiStripper = ansi.Strip
