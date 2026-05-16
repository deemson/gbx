package tui2

import (
	tea "charm.land/bubbletea/v2"
)

type tuiModel struct {
	repos reposModel
}

func newModel(dir string) tuiModel {
	return tuiModel{
		repos: newReposModel(dir),
	}
}

func (m tuiModel) Init() tea.Cmd {
	return m.repos.Init()
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.repos = m.repos.SetWidth(msg.Width)
		return m, nil
	}
	var cmd tea.Cmd
	m.repos, cmd = m.repos.Update(msg)
	return m, cmd
}

func (m tuiModel) View() tea.View {
	return tea.View{
		Content:   m.repos.View(),
		AltScreen: true,
	}
}
