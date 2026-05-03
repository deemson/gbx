package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/tui/repos"
)

type model struct {
	repos repos.Model
}

func newModel() model {
	return model{
		repos: repos.NewModel(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		repos.InitCmd,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	var reposCmd tea.Cmd
	m.repos, reposCmd = m.repos.Update(msg)
	return m, tea.Batch(reposCmd)
}

func (m model) View() tea.View {
	return tea.View{
		Content:   m.repos.View(),
		AltScreen: true,
	}
}
