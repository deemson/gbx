package tui2

type config struct {
	dir string
}

type Option func(*config)

func WithDir(dir string) Option {
	return func(c *config) { c.dir = dir }
}
