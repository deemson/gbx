package tui

import (
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/davecgh/go-spew/spew"
)

type Main struct {
	altScreen bool
	msgs      []tea.Msg
}

func (m Main) Init() tea.Cmd {
	return nil
}

func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.msgs = append(m.msgs, msg)
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "a":
			m.altScreen = !m.altScreen
		case "l":
			cmd := exec.Command("lazygit", "-p", ".")
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				if err != nil {
					return tea.Quit
				}
				return nil
			})
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Main) View() tea.View {
	return tea.View{
		Content:     spew.Sdump(m.msgs),
		WindowTitle: "gbx",
		AltScreen:   m.altScreen,
	}
}

func Run() error {
	_, err := tea.NewProgram(Main{}).Run()
	return err
}
