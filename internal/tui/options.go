package tui

type config struct {
	dir     string
	version string
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
