package row

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/tui/repos/gitreport"
	"github.com/rs/zerolog/log"
)

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
			log.Error().
				Err(err).
				Str("name", m.name).
				Msg("failed loading status")
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
			if errors.Is(err, git.ErrRepositoryHasNoCommits) {
				return LinesChangedMsg{
					msg: msg{
						Name: m.name,
					},
					LinesChanged: gitreport.LinesChanged{
						Added:   0,
						Deleted: 0,
					},
				}
			}
			log.Error().
				Err(err).
				Str("name", m.name).
				Msg("failed loading lines changed")
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

func (m Model) View() []string {
	status := "..."
	if m.status != nil {
		status = fmt.Sprintf("%s +%d -%d", m.status.Branch, m.status.Ahead, m.status.Behind)
	}
	diff := "..."
	if m.linesChanged != nil {
		diff = fmt.Sprintf("+%d -%d", m.linesChanged.Added, m.linesChanged.Deleted)
	}
	return []string{
		m.name,
		status,
		diff,
	}
}
