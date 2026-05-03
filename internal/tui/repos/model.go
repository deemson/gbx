package repos

import (
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/tui/repos/row"
	"github.com/rs/zerolog/log"
)

type Model struct {
	directory      string
	rowsByRepoName map[string]row.Model
	table          table.Model
}

func NewModel() Model {
	tbl := table.New(
		table.WithColumns(row.TableColumns()),
		table.WithFocused(false),
		table.WithWidth(50),
	)
	return Model{
		directory:      "",
		rowsByRepoName: map[string]row.Model{},
		table:          tbl,
	}
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
		m.refreshTable()
		return m, r.Refresh()
	case row.Msg:
		var cmd tea.Cmd
		m.rowsByRepoName[msg.RepoName()], cmd = m.rowsByRepoName[msg.RepoName()].Update(msg)
		m.refreshTable()
		return m, cmd
	}
	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)
	return m, tea.Batch(tableCmd)
}

func (m *Model) refreshTable() {
	rows := make([]table.Row, 0, len(m.rowsByRepoName))
	for _, row := range m.rowsByRepoName {
		rows = append(rows, row.TableRow())
	}
	m.table.SetRows(rows)
}

func (m Model) View() string {
	if m.directory == "" {
		return "loading directory"
	}
	if len(m.rowsByRepoName) == 0 {
		return "discovering repos"
	}
	return m.table.View()
}
