package tui2

import (
	"errors"

	tea "charm.land/bubbletea/v2"
)

func Run(opts ...Option) error {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.dir == "" {
		return errors.New("tui2: WithDir is required")
	}
	program := tea.NewProgram(newModel(cfg.dir))
	_, err := program.Run()
	return err
}
