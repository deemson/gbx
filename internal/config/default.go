package config

func Default() Config {
	return Config{
		Actions: []Action{
			{Label: "lazygit", Command: []string{"lazygit"}},
			{Label: "shell", Command: []string{"{{ env.SHELL }}"}},
		},
	}
}
