package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modeNormal mode = iota
	modeCommand
)

type model struct {
	tabs         []string
	activeTabIndex int
	curMode      mode
	terminal     string
}

func InitialModel() model {
	return model{
		tabs:           []string{"Session 1", "Session 2", "Session 3"},
		activeTabIndex: 0,
		curMode:        modeNormal,
		terminal:       "Terminal content goes here...",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.curMode {
		case modeNormal:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case ":":
				m.curMode = modeCommand
			case "tab":
				m.activeTabIndex = (m.activeTabIndex + 1) % len(m.tabs)
			}
		case modeCommand:
			switch msg.String() {
			case "esc", "enter":
				m.curMode = modeNormal
			}
		}
	}

	return m, nil
}

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
)

func (m model) View() string {
	doc := strings.Builder{}

	// Tabs
	var renderedTabs []string
	for i, t := range m.tabs {
		style := tabStyle
		if i == m.activeTabIndex {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...) + "\n")

	// Terminal Placeholder
	doc.WriteString(fmt.Sprintf("\n  Selected: %s\n\n", m.tabs[m.activeTabIndex]))
	doc.WriteString("  " + m.terminal + "\n\n")

	// Status Bar
	modeStr := " NORMAL "
	style := statusStyle
	if m.curMode == modeCommand {
		modeStr = " COMMAND "
		style = commandModeStyle
	}
	
	doc.WriteString(style.Render(modeStr))
	if m.curMode == modeCommand {
		doc.WriteString(" :_")
	} else {
		doc.WriteString(" (Press : for command mode, Tab to switch sessions)")
	}

	return doc.String()
}
