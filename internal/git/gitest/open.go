package gitest

import (
	"context"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/stretchr/testify/require"
)

func Open(t *testing.T, path string) Repo {
	ctx := context.Background()
	repo, err := git.Open(ctx, path)
	require.NoError(t, err)
	return Repo{
		repo: repo,
		t:    t,
	}
}
