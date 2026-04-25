package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/tui/model"
)

func Run() error {
	_, err := tea.NewProgram(model.Main{}).Run()
	return err
}
