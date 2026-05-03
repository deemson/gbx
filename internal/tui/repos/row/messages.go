package row

import "github.com/deemson/gbx/internal/tui/repos/gitreport"

type Msg interface {
	RepoName() string
}

type msg struct {
	Name string
}

func (m msg) RepoName() string {
	return m.Name
}

type StatusMsg struct {
	msg
	Status gitreport.Status
}

type LinesChangedMsg struct {
	msg
	LinesChanged gitreport.LinesChanged
}
