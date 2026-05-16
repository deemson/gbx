package repos

import (
	"sort"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/deemson/gbx/internal/tui/repos/row"
	"github.com/rs/zerolog/log"
)

const (
	initialWidth = 10
)

type Model struct {
	directory      string
	rowsByRepoName map[string]row.Model
	width          int
	table          *table.Table
}

func NewModel() Model {
	return Model{
		directory:      "",
		rowsByRepoName: map[string]row.Model{},
		width:          initialWidth,
		table:          table.New().Border(lipgloss.HiddenBorder()).Width(initialWidth),
	}
}

func (m Model) SetWidth(width int) Model {
	m.width = width
	m.table = m.table.Width(width)
	return m
}

func (m Model) refreshTableRows() Model {
	names := make([]string, 0, len(m.rowsByRepoName))
	for name := range m.rowsByRepoName {
		names = append(names, name)
	}
	sort.Strings(names)
	rows := make([][]string, len(names))
	for i, name := range names {
		rows[i] = m.rowsByRepoName[name].View()
	}
	m.table = m.table.ClearRows().Rows(rows...)
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case InitMsg:
		log.Debug().Str("directory", msg.Directory).Msg("init started")
		m.directory = msg.Directory
		openRepoCmds := make([]tea.Cmd, len(msg.DirEntries))
		for i, dirEntry := range msg.DirEntries {
			openRepoCmds[i] = newOpenRepoCmd(m.directory, dirEntry)
		}
		return m, tea.Sequence(
			tea.Batch(openRepoCmds...),
			initDoneCmd,
		)
	case InitDoneMsg:
		return m, nil
	case RepoFoundMsg:
		log.Debug().
			Str("name", msg.Name).
			Str("path", msg.Repo.Path()).
			Msg("found repo")
		r := row.NewModel(msg.Name, msg.Repo)
		m.rowsByRepoName[msg.Name] = r
		m = m.refreshTableRows()
		return m, r.Refresh()
	case row.Msg:
		var cmd tea.Cmd
		m.rowsByRepoName[msg.RepoName()], cmd = m.rowsByRepoName[msg.RepoName()].Update(msg)
		m = m.refreshTableRows()
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if m.directory == "" {
		return "loading directory"
	}
	if len(m.rowsByRepoName) == 0 {
		return "discovering repos"
	}
	return m.table.Render()
}
