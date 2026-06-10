package config

func Default() Config {
	return Config{
		Actions: Actions{
			Enter:      []string{"lazygit"},
			ShiftEnter: []string{"{ env.SHELL }"},
		},
	}
}
