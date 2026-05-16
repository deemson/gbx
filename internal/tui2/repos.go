package tui2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/gitreport"
	"github.com/rs/zerolog/log"
)

const initialWidth = 10

type row struct {
	name         string
	repo         git.Repo
	status       *gitreport.Status
	linesChanged *gitreport.LinesChanged
}

func (r row) cells() []string {
	status := "..."
	if r.status != nil {
		status = fmt.Sprintf("%s +%d -%d", r.status.Branch, r.status.Ahead, r.status.Behind)
	}
	diff := "..."
	if r.linesChanged != nil {
		diff = fmt.Sprintf("+%d -%d", r.linesChanged.Added, r.linesChanged.Deleted)
	}
	return []string{r.name, status, diff}
}

type reposModel struct {
	directory string
	rows      map[string]row
	width     int
	table     *table.Table
}

func newReposModel(directory string) reposModel {
	return reposModel{
		directory: directory,
		rows:      map[string]row{},
		width:     initialWidth,
		table:     table.New().Border(lipgloss.HiddenBorder()).Width(initialWidth),
	}
}

func (m reposModel) Init() tea.Cmd {
	return readEntriesCmd(m.directory)
}

func (m reposModel) SetWidth(width int) reposModel {
	m.width = width
	m.table = m.table.Width(width)
	return m
}

func (m reposModel) refreshTableRows() reposModel {
	names := make([]string, 0, len(m.rows))
	for name := range m.rows {
		names = append(names, name)
	}
	sort.Strings(names)
	cells := make([][]string, len(names))
	for i, name := range names {
		cells[i] = m.rows[name].cells()
	}
	m.table = m.table.ClearRows().Rows(cells...)
	return m
}

func (m reposModel) Update(msg tea.Msg) (reposModel, tea.Cmd) {
	switch msg := msg.(type) {
	case entriesLoadedMsg:
		log.Debug().Str("directory", m.directory).Int("entries", len(msg.entries)).Msg("entries loaded")
		cmds := make([]tea.Cmd, len(msg.entries))
		for i, e := range msg.entries {
			cmds[i] = openRepoCmd(m.directory, e)
		}
		return m, tea.Batch(cmds...)
	case repoFoundMsg:
		log.Debug().Str("name", msg.name).Str("path", msg.repo.Path()).Msg("found repo")
		m.rows[msg.name] = row{name: msg.name, repo: msg.repo}
		m = m.refreshTableRows()
		return m, tea.Batch(
			statusCmd(msg.name, msg.repo),
			linesChangedCmd(msg.name, msg.repo),
		)
	case statusLoadedMsg:
		log.Debug().Str("name", msg.name).Msg("repo status loaded")
		r := m.rows[msg.name]
		r.status = &msg.status
		m.rows[msg.name] = r
		m = m.refreshTableRows()
	case linesChangedLoadedMsg:
		log.Debug().Str("name", msg.name).Msg("repo diff loaded")
		r := m.rows[msg.name]
		r.linesChanged = &msg.linesChanged
		m.rows[msg.name] = r
		m = m.refreshTableRows()
	}
	return m, nil
}

func (m reposModel) View() string {
	if len(m.rows) == 0 {
		return "discovering repos"
	}
	return m.table.Render()
}

type entriesLoadedMsg struct {
	entries []os.DirEntry
}

type repoFoundMsg struct {
	name string
	repo git.Repo
}

type statusLoadedMsg struct {
	name   string
	status gitreport.Status
}

type linesChangedLoadedMsg struct {
	name         string
	linesChanged gitreport.LinesChanged
}

func readEntriesCmd(directory string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(directory)
		if err != nil {
			log.Error().Err(err).Str("directory", directory).Msg("failed to read directory")
			return nil
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func openRepoCmd(dir string, entry os.DirEntry) tea.Cmd {
	return func() tea.Msg {
		if !entry.IsDir() {
			return nil
		}
		repo, err := git.Open(context.Background(), filepath.Join(dir, entry.Name()))
		if err != nil {
			if errors.Is(err, git.ErrNotRepository) {
				return nil
			}
			log.Error().Err(err).Str("dir", dir).Str("entry", entry.Name()).Msg("failed to open repo")
			return nil
		}
		return repoFoundMsg{name: entry.Name(), repo: repo}
	}
}

func statusCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		status, err := repo.Status(context.Background())
		if err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed loading status")
			return nil
		}
		return statusLoadedMsg{
			name:   name,
			status: gitreport.NewStatus(context.Background(), status),
		}
	}
}

func linesChangedCmd(name string, repo git.Repo) tea.Cmd {
	return func() tea.Msg {
		diffNumStat, err := repo.DiffNumStatHead(context.Background())
		if err != nil {
			if errors.Is(err, git.ErrRepositoryHasNoCommits) {
				return linesChangedLoadedMsg{name: name, linesChanged: gitreport.LinesChanged{}}
			}
			log.Error().Err(err).Str("name", name).Msg("failed loading lines changed")
			return nil
		}
		return linesChangedLoadedMsg{
			name:         name,
			linesChanged: gitreport.NewLinesChanged(diffNumStat),
		}
	}
}
