package tui

import appconfig "github.com/deemson/gbx/internal/config"

type config struct {
	dirs      []Directory
	version   string
	logPath   string
	appConfig appconfig.Config
}

// Directory is one directory-list occurrence. Label is rendered exactly as
// supplied by the caller unless HideHeading is set; Path is scanned on disk.
type Directory struct {
	Label       string
	Path        string
	HideHeading bool
}

type Option func(*config)

// WithDir sets the root directory whose immediate subdirectories are scanned
// for git repositories.
func WithDir(dir string) Option {
	return WithDirectories([]Directory{{Label: dir, Path: dir}})
}

// WithDirectories sets the ordered directory occurrences whose immediate
// subdirectories are scanned for git repositories.
func WithDirectories(dirs []Directory) Option {
	return func(c *config) { c.dirs = append([]Directory(nil), dirs...) }
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
// enter-key menu lists its actions and runs the chosen one in the cursored repo.
func WithConfig(cfg appconfig.Config) Option {
	return func(c *config) { c.appConfig = cfg }
}
