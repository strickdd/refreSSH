package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	doc := strings.Builder{}

	// Tabs
	var renderedTabs []string
	for i, t := range m.tabs {
		style := tabStyle
		if i == m.activeTabIndex {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(fmt.Sprintf("%d:%s", i+1, t)))
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(tabsRow + "\n")

	// Terminal area
	contentHeight := m.height - 4 // Tabs + Status bar + Borders
	if contentHeight < 0 {
		contentHeight = 0
	}
	
	terminalContent := m.terminal
	if m.activeTabIndex < len(m.tabs) {
		terminalContent = fmt.Sprintf("Session: %s\n\n%s", m.tabs[m.activeTabIndex], m.terminal)
	}

	window := windowStyle.
		Width(m.width - 2).
		Height(contentHeight).
		Render(terminalContent)
	doc.WriteString(window + "\n")

	// Status Bar
	modeStr := " NORMAL "
	style := statusStyle
	if m.dispatcher.InCommand() {
		modeStr = " COMMAND "
		style = commandModeStyle
	}

	status := style.Render(modeStr)
	if m.dispatcher.InCommand() {
		status += " Waiting for command..."
	} else {
		status += fmt.Sprintf(" %s - Tab %d/%d", m.dispatcher.prefix, m.activeTabIndex+1, len(m.tabs))
	}

	doc.WriteString(status)

	return doc.String()
}
