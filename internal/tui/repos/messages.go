package repos

import (
	"os"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/tui/repos/gitreport"
)

type InitMsg struct {
	Directory  string
	DirEntries []os.DirEntry
}

type RepoFoundMsg struct {
	Name string
	Repo git.Repo
}

type RepoStatusMsg struct {
	Name   string
	Status gitreport.Status
}

type RepoLinesChangedMsg struct {
	Name         string
	LinesChanged gitreport.LinesChanged
}

type InitDoneMsg struct{}
