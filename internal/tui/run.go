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
	m := newModel(cfg.dir)
	if cfg.version != "" {
		m.version = cfg.version
	}
	_, err := tea.NewProgram(m).Run()
	return err
}
