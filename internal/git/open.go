package git

import (
	"context"
	"errors"
	"strings"

	"github.com/deemson/gbx/internal/git/gitexec"
)

var (
	ErrDoesNotExist  = errors.New("path does not exist")
	ErrNotDirectory  = errors.New("not a directory")
	ErrNotRepository = errors.New("not a git repository")
)

func Open(ctx context.Context, path string) (Repo, error) {
	res, err := gitexec.Run(ctx, path, "rev-parse", "--show-toplevel")
	if err != nil {
		if res.ExitCode == 128 {
			switch {
			case strings.Contains(res.Stderr, "not a git repository"):
				return Repo{}, ErrNotRepository
			case strings.Contains(res.Stderr, "Not a directory"):
				return Repo{}, ErrNotDirectory
			case strings.Contains(res.Stderr, "No such file or directory"):
				return Repo{}, ErrDoesNotExist
			}
		}
		return Repo{}, err
	}
	return Repo{path: path}, nil
}
