package git

import (
	"context"
	"errors"
	"strings"

	"github.com/deemson/gbx/internal/git/exec"
)

func Init(ctx context.Context, path string) (Repo, error) {
	res, err := exec.Git{
		Path: path,
	}.Run(ctx, "init")
	if err != nil {
		return Repo{}, NewErrUnknown(res, err)
	}
	if !strings.Contains(string(res.Stdout), "Initialized empty Git repository") {
		return Repo{}, NewErrUnknown(res, errors.New("unexpecte stdout"))
	}
	return Repo{
		path: path,
	}, nil
}
