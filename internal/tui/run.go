package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	InitialURL    string
	InitialMode   string
	InitialOutput string
}

func Run(config Config) error {
	p := tea.NewProgram(NewModel(config), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui execution failed: %w", err)
	}
	return nil
}
