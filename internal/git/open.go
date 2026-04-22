package git

import (
	"context"
	"strings"

	"github.com/deemson/gbx/internal/git/exec"
)

func Open(ctx context.Context, path string) (Repo, error) {
	res, err := exec.Git{
		Path: path,
	}.Run(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		if res.ExitCode == 128 {
			stderr := string(res.Stderr)
			switch {
			case strings.Contains(stderr, "not a git repository"):
				return Repo{}, ErrNotRepository
			case strings.Contains(stderr, "Not a directory"):
				return Repo{}, ErrNotDirectory
			case strings.Contains(stderr, "No such file or directory"):
				return Repo{}, ErrDoesNotExist
			}
		}
		return Repo{}, NewErrUnknown(res, err)
	}
	return Repo{path: path}, nil
}
