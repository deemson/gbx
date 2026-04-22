package git

import (
	"context"

	"github.com/deemson/gbx/internal/git/exec"
)

type Repo struct {
	path string
}

func (r Repo) Path() string {
	return r.path
}

func (r Repo) Status(ctx context.Context) (any, error) {
	res, err := exec.Git{
		Path: r.path,
	}.Run(ctx, "status", "--null", "--porcelain=2")
	return nil, NewErrUnknown(res, err)
}
