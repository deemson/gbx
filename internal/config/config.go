package config

type Config struct {
	Actions Actions `toml:"actions" gozod:"required"`
}

type Actions struct {
	Enter      []string `toml:"enter" gozod:"required"`
	ShiftEnter []string `toml:"shift-enter" gozod:"required"`
}


