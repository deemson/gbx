package config

type Config struct {
	Actions []Action `toml:"actions" gozod:"required,min=1,max=9"`
}

// Action is one entry in the enter-key menu: Label is the text shown next to
// its digit, Command is the arg vector run (after {{ }} interpolation) in the
// cursored repo's directory.
type Action struct {
	Label   string   `toml:"label" gozod:"required"`
	Command []string `toml:"command" gozod:"required,min=1"`
}
