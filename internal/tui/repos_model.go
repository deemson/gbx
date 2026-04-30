package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/deemson/gbx/internal/git"
)

type reposModel struct {
	table   table.Model
	loading bool
	err     error
}

func newReposModel() reposModel {
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 20},
			{Title: "Branch", Width: 20},
			{Title: "Lines", Width: 16},
		}),
		table.WithFocused(true),
	)
	return reposModel{
		table:   t,
		loading: true,
	}
}

func (m reposModel) Init() tea.Cmd {
	return loadReposCmd()
}

type reposLoadedMsg struct {
	rows []table.Row
	cols []table.Column
	err  error
}

func loadReposCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cwd, err := os.Getwd()
		if err != nil {
			return reposLoadedMsg{err: fmt.Errorf("getwd: %w", err)}
		}

		entries, err := os.ReadDir(cwd)
		if err != nil {
			return reposLoadedMsg{err: fmt.Errorf("readdir %s: %w", cwd, err)}
		}

		type repoResult struct {
			name    string
			branch  string
			added   int
			deleted int
			ok      bool
		}

		var (
			results []repoResult
			mu      sync.Mutex
			wg      sync.WaitGroup
			sem     = make(chan struct{}, 8)
		)

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(cwd, e.Name())
			repo, err := git.Open(ctx, path)
			if err != nil {
				if errors.Is(err, git.ErrNotRepository) ||
					errors.Is(err, git.ErrNotDirectory) ||
					errors.Is(err, git.ErrDoesNotExist) {
					continue
				}
				mu.Lock()
				results = append(results, repoResult{name: e.Name()})
				mu.Unlock()
				continue
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(name string, repo git.Repo) {
				defer wg.Done()
				defer func() { <-sem }()

				res := repoResult{name: name, ok: true}

				status, sErr := repo.Status(ctx)
				if sErr != nil {
					res.ok = false
				} else {
					res.branch = status.Branch
					if res.branch == "" {
						res.branch = "(detached)"
					}
				}

				diff, dErr := repo.DiffNumStatHead(ctx)
				if dErr != nil {
					res.ok = false
				} else {
					for _, p := range diff.Paths {
						res.added += p.AddedLines
						res.deleted += p.DeletedLines
					}
				}

				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}(e.Name(), repo)
		}

		wg.Wait()

		sort.Slice(results, func(i, j int) bool {
			return results[i].name < results[j].name
		})

		rows := make([]table.Row, 0, len(results))
		nameW, branchW := len("Name"), len("Branch")
		for _, r := range results {
			var branchCell, linesCell string
			if r.ok {
				branchCell = r.branch
				linesCell = fmt.Sprintf("+%d -%d", r.added, r.deleted)
			} else {
				branchCell = "!"
				linesCell = "!"
			}
			rows = append(rows, table.Row{r.name, branchCell, linesCell})
			if l := len(r.name); l > nameW {
				nameW = l
			}
			if l := len(branchCell); l > branchW {
				branchW = l
			}
		}

		cols := []table.Column{
			{Title: "Name", Width: nameW + 2},
			{Title: "Branch", Width: branchW + 2},
			{Title: "Lines", Width: 18},
		}

		return reposLoadedMsg{rows: rows, cols: cols}
	}
}

func (m reposModel) Update(msg tea.Msg) (reposModel, tea.Cmd) {
	switch msg := msg.(type) {
	case reposLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.table.SetColumns(msg.cols)
		m.table.SetRows(msg.rows)
		return m, nil
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		h := msg.Height - 4
		if h < 3 {
			h = 3
		}
		m.table.SetHeight(h)
		return m, nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m reposModel) View() string {
	if m.loading {
		return "Loading repos..."
	}
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	return m.table.View()
}
