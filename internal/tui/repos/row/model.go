package row

import (
	"context"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/tui/repos/gitreport"
	"github.com/rs/zerolog/log"
)

func TableColumns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 10},
		{Title: "Branch", Width: 10},
		{Title: "Commit", Width: 10},
		{Title: "Status", Width: 10},
		{Title: "Diff", Width: 10},
	}
}

type Model struct {
	name         string
	repo         git.Repo
	status       *gitreport.Status
	linesChanged *gitreport.LinesChanged
}

func NewModel(name string, repo git.Repo) Model {
	return Model{
		name:         name,
		repo:         repo,
		status:       nil,
		linesChanged: nil,
	}
}

func (m Model) Status() tea.Cmd {
	return func() tea.Msg {
		status, err := m.repo.Status(context.Background())
		if err != nil {
		}
		return StatusMsg{
			msg: msg{
				Name: m.name,
			},
			Status: gitreport.NewStatus(context.Background(), status),
		}
	}
}

func (m Model) LinesChanged() tea.Cmd {
	return func() tea.Msg {
		diffNumStat, err := m.repo.DiffNumStatHead(context.Background())
		if err != nil {
		}
		return LinesChangedMsg{
			msg: msg{
				Name: m.name,
			},
			LinesChanged: gitreport.NewLinesChanged(diffNumStat),
		}
	}
}

func (m Model) Refresh() tea.Cmd {
	return tea.Batch(m.Status(), m.LinesChanged())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusMsg:
		log.Debug().
			Str("name", msg.Name).
			Msg("repo status loaded")
		m.status = &msg.Status
	case LinesChangedMsg:
		log.Debug().
			Str("name", msg.Name).
			Msg("repo diff loaded")
		m.linesChanged = &msg.LinesChanged
	}
	return m, nil
}

func (m Model) TableRow() table.Row {
	branch, commit, status := m.renderStatus()
	linesChanged := m.renderLinesChanged()
	return table.Row{
		m.name,
		branch,
		commit,
		status,
		linesChanged,
	}
}

func (m Model) renderStatus() (string, string, string) {
	if m.status == nil {
		return "loading status", "", ""
	}
	return m.status.Branch, m.status.Commit, "ok"
}

func (m Model) renderLinesChanged() string {
	if m.linesChanged == nil {
		return "loading lines"
	}
	return "ok"
}
