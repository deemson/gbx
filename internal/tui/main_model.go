package tui

import tea "charm.land/bubbletea/v2"

type mainModel struct {
	repos reposModel
}

func newMainModel() mainModel {
	return mainModel{repos: newReposModel()}
}

func (m mainModel) Init() tea.Cmd {
	return m.repos.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.repos, cmd = m.repos.Update(msg)
	return m, cmd
}

func (m mainModel) View() tea.View {
	return tea.View{
		Content: m.repos.View(),
	}
}
