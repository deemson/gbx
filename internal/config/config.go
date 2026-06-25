package config

type Config struct {
	Actions []Action `toml:"actions"`
}

// Action is one entry in the enter-key menu: Label is the text shown next to
// its digit, Command is the arg vector run (after {{ }} interpolation) in the
// cursored repo's directory.
type Action struct {
	Label   string   `toml:"label"`
	Command []string `toml:"command"`
}
