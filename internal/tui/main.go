package tui

import tea "charm.land/bubbletea/v2"

type Main struct {
}

func (m Main) Init() tea.Cmd {
	return nil
}

func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Main) View() tea.View {
	return tea.View{
		Content:     "hello world",
		WindowTitle: "sup",
		AltScreen:   true,
	}
}

func Run() {
	if _, err := tea.NewProgram(Main{}).Run(); err != nil {
		panic(err.Error())
	}
}
