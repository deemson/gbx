package tui

import (
	"errors"

	tea "charm.land/bubbletea/v2"
)

// Run starts the TUI and blocks until it exits.
func Run(opts ...Option) error {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.dir == "" {
		return errors.New("tui: WithDir is required")
	}
	_, err := tea.NewProgram(newModel(cfg.dir)).Run()
	return err
}
