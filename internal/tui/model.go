package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// model is the root TUI model. The filter input is always focused (fzf-style):
// printable keys edit the filter, and every action is a non-printable binding.
type model struct {
	dir    string
	filter textinput.Model
	width  int
	height int
}

func newModel(dir string) model {
	filter := textinput.New()
	filter.Prompt = "> "
	filter.Placeholder = "filter repos"
	return model{
		dir:    dir,
		filter: filter,
	}
}

func (m model) Init() tea.Cmd {
	return m.filter.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	// Discovery and the repo table arrive in later slices; for now the body is
	// just a placeholder beneath the always-focused filter.
	body := "no repos"
	return tea.View{
		Content:   m.filter.View() + "\n\n" + body,
		AltScreen: true,
	}
}
