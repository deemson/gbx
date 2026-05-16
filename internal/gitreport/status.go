package gitreport

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/deemson/gbx/internal/git"
	"github.com/rs/zerolog"
)

type Status struct {
	Branch        string
	Commit        string
	Upstream      string
	Ahead         int
	Behind        int
	Unknown       int
	Untracked     int
	Conflicts     int
	Added         int
	ModifiedIndex int
	ModifiedFS    int
	Moved         int
	DeletedIndex  int
	DeletedFS     int
}

func NewStatus(ctx context.Context, status git.Status) Status {
	logger := zerolog.Ctx(ctx).With().
		Str("branch", status.Branch).
		Str("commit", status.Commit).
		Logger()

	report := Status{
		Branch:   status.Branch,
		Commit:   status.Commit,
		Upstream: status.Upstream,
		Ahead:    status.Ahead,
		Behind:   status.Behind,
	}

	for _, pathStatus := range status.Paths {
		switch pathStatus := pathStatus.(type) {
		case git.UntrackedPathStatus:
			report.Untracked += 1
		case git.ConflictPathStatus:
			report.Conflicts += 1
		case git.RegularPathStatus:
			stateIndex := pathStatus.StateIndex
			stateFS := pathStatus.StateFS
			pathLogger := logger.With().
				Str("status", "regular").
				Any("index_state", stateIndex).
				Any("fs_state", stateFS).
				Str("path", pathStatus.Path).
				Logger()
			switch {
			case stateIndex == git.AddedPathState && stateFS == git.NotChangedPathState:
				report.Added += 1
			case stateIndex == git.ModifiedPathState && stateFS == git.NotChangedPathState:
				report.ModifiedIndex += 1
			case stateIndex == git.NotChangedPathState && stateFS == git.ModifiedPathState:
				report.ModifiedFS += 1
			case stateIndex == git.DeletedPathState && stateFS == git.NotChangedPathState:
				report.DeletedIndex += 1
			case stateIndex == git.NotChangedPathState && stateFS == git.DeletedPathState:
				report.DeletedFS += 1
			default:
				pathLogger.Warn().Msg("unknown combination of path states")
				report.Unknown += 1
			}
		case git.MovedPathStatus:
			stateIndex := pathStatus.StateIndex
			stateFS := pathStatus.StateFS
			pathLogger := logger.With().
				Str("status", "moved").
				Any("index_state", stateIndex).
				Any("fs_state", stateFS).
				Str("path", pathStatus.Path).
				Logger()
			switch {
			case stateIndex == git.RenamedPathState && stateFS == git.NotChangedPathState:
				report.Moved += 1
			default:
				pathLogger.Warn().Msg("unknown combination of path states")
				report.Unknown += 1
			}
		default:
			logger.Warn().Str("status", spew.Sdump(pathStatus)).Msg("unknown path status")
			report.Unknown += 1
		}
	}

	return report
}
