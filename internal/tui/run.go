package tui

import (
	tea "charm.land/bubbletea/v2"
)

func Run() error {
	program := tea.NewProgram(newModel())
	_, err := program.Run()
	return err
}
