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
	Run: func(_ *cobra.Command, _ []string) {
		m := tui.InitialModel()
		if m.Err() != nil {
			fmt.Printf("Error: %v\n", m.Err())
			return
		}

		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting TUI: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
