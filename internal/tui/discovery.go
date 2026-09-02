package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog/log"
)

type entriesLoadedMsg struct {
	section int
	entries []os.DirEntry
}

type repoCheckedMsg struct {
	section int
	name    string
	repo    git.Repo
	found   bool
}

// repoFoundMsg is retained for direct model tests; discovery itself reports a
// repoCheckedMsg so a section can also count non-repository children.
type repoFoundMsg struct {
	name string
	repo git.Repo
}

// readEntriesCmd lists the immediate entries of the root directory.
func readEntriesCmd(section int, dir string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Error().Err(err).Str("dir", dir).Msg("failed to read directory")
			return entriesLoadedMsg{section: section}
		}
		return entriesLoadedMsg{section: section, entries: entries}
	}
}

// openRepoCmd tries to open one entry as a git repository. Non-directories and
// non-repositories are silently skipped; only real open failures are logged.
func openRepoCmd(section int, dir string, entry os.DirEntry) tea.Cmd {
	return func() tea.Msg {
		if !entry.IsDir() {
			return repoCheckedMsg{section: section}
		}
		repo, err := git.Open(context.Background(), filepath.Join(dir, entry.Name()))
		if err != nil {
			if !errors.Is(err, git.ErrNotRepository) {
				log.Error().Err(err).Str("dir", dir).Str("name", entry.Name()).Msg("failed to open repo")
			}
			return repoCheckedMsg{section: section}
		}
		// Opened only because it sits inside an enclosing repo (its toplevel is
		// elsewhere) — not its own root, so it's not a listable repo. Silent skip,
		// like the non-repo case.
		if repo.Root() != repo.Path() {
			return repoCheckedMsg{section: section}
		}
		return repoCheckedMsg{section: section, name: entry.Name(), repo: repo, found: true}
	}
}
