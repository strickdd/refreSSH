package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the TUI interface",
	RunE: func(_ *cobra.Command, _ []string) error {
		m := tui.InitialModel()
		if err := m.Err(); err != nil {
			return fmt.Errorf("failed to initialize TUI: %w", err)
		}

		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI exited with error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
