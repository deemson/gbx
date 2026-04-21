package git

import (
	"context"

	"github.com/deemson/gbx/internal/git/gitexec"
)

type Repo struct {
	path string
}

func (r Repo) Path() string {
	return r.path
}

func (r Repo) Status(ctx context.Context) error {
	_, err := gitexec.Run(ctx, r.path, "status", "--porcelain")
	return err
}
