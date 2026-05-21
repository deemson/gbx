package tui

type config struct {
	dir string
}

type Option func(*config)

// WithDir sets the root directory whose immediate subdirectories are scanned
// for git repositories.
func WithDir(dir string) Option {
	return func(c *config) { c.dir = dir }
}
