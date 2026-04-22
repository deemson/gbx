package gitest

import (
	"context"

	"github.com/deemson/gbx/internal/git"
)

func Init(ctx context.Context, path string) (Repo, error) {
	repo, err := git.Init(ctx, path)
	if err != nil {
		return Repo{}, err
	}
	return Repo{
		Repo: repo,
	}, nil
}
