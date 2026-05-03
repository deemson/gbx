package repos

import (
	"context"
	"errors"
	"os"
	"path"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git"
)

func InitCmd() tea.Msg {
	directory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dirEntries, err := os.ReadDir(directory)
	if err != nil {
		panic(err)
	}
	return InitMsg{
		Directory:  directory,
		DirEntries: dirEntries,
	}
}

func newOpenRepoCmd(dir string, dirEntry os.DirEntry) tea.Cmd {
	return func() tea.Msg {
		if !dirEntry.IsDir() {
			return nil
		}
		repo, err := git.Open(context.Background(), path.Join(dir, dirEntry.Name()))
		if err != nil {
			if errors.Is(err, git.ErrNotRepository) {
				return nil
			}
			panic(err)
		}
		return RepoFoundMsg{
			Name: dirEntry.Name(),
			Repo: repo,
		}
	}
}

func initDoneCmd() tea.Msg {
	return InitDoneMsg{}
}
