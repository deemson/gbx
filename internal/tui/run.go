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
	if len(cfg.dirs) == 0 {
		return errors.New("tui: WithDir or WithDirectories is required")
	}
	m := newModelWithDirectories(cfg.dirs)
	if cfg.version != "" {
		m.version = cfg.version
	}
	m.logPath = cfg.logPath
	m.appConfig = cfg.appConfig
	_, err := tea.NewProgram(m).Run()
	return err
}
