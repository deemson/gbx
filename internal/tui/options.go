package tui

import appconfig "github.com/deemson/gbx/internal/config"

type config struct {
	dir       string
	version   string
	logPath   string
	appConfig appconfig.Config
}

type Option func(*config)

// WithDir sets the root directory whose immediate subdirectories are scanned
// for git repositories.
func WithDir(dir string) Option {
	return func(c *config) { c.dir = dir }
}

// WithVersion sets the version string shown in the header's right corner. Empty
// (e.g. a plain `go build` with no ldflags) leaves the model's "dev" default.
func WithVersion(version string) Option {
	return func(c *config) { c.version = version }
}

// WithLogPath sets the log file path shown in the help overlay. main.go owns the
// canonical path (xdg.StateFile), so it passes the same value it writes to.
func WithLogPath(path string) Option {
	return func(c *config) { c.logPath = path }
}

// WithConfig sets the loaded application config (the action command set). The
// model stores it; it is not yet consumed by any key handler.
func WithConfig(cfg appconfig.Config) Option {
	return func(c *config) { c.appConfig = cfg }
}
