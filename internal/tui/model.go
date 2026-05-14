package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/deemson/gbx/internal/tui/repos"
)

type model struct {
	repos repos.Model
	input textinput.Model
}

func newModel() model {
	input := textinput.New()
	// input.SetVirtualCursor(false)
	input.ShowSuggestions = true
	input.SetSuggestions([]string{"thing1", "thing2"})
	// input.Placeholder = "something"
	input.SetWidth(15)
	input.Focus()
	return model{
		repos: repos.NewModel(),
		input: input,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		repos.InitCmd,
		textinput.Blink,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	}
	var reposCmd tea.Cmd
	m.repos, reposCmd = m.repos.Update(msg)
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	return m, tea.Batch(reposCmd, inputCmd)
}

func (m *model) resize(width, height int) {
	m.input.SetWidth(width)
	m.repos = m.repos.Resize(width, max(0, height-lipgloss.Height(m.input.View())))
}

func (m model) View() tea.View {
	repos := m.repos.View()
	// cursor := m.input.Cursor()
	// cursor.Y = lipgloss.Height(repos)
	return tea.View{
		Content: lipgloss.JoinVertical(
			lipgloss.Top,
			repos,
			m.input.View(),
		),
		AltScreen: true,
		// Cursor:    cursor,
	}
}
