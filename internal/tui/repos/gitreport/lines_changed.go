package gitreport

import "github.com/deemson/gbx/internal/git"

type LinesChanged struct {
	Added   int
	Deleted int
}

func NewLinesChanged(diffNumStat git.DiffNumStat) LinesChanged {
	added := 0
	deleted := 0
	for _, pathDiffNumStat := range diffNumStat.Paths {
		added += pathDiffNumStat.AddedLines
		deleted += pathDiffNumStat.DeletedLines
	}
	return LinesChanged{
		Added:   added,
		Deleted: deleted,
	}
}
