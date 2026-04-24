package gitest

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/exec"
	"github.com/stretchr/testify/require"
)

func Init(t *testing.T, path string) Repo {
	ctx := context.Background()
	res, err := exec.Git{
		Path: path,
	}.Run(ctx, "init")
	if err != nil {
		require.NoError(t, git.NewUnknownRunErr(res, err))
	}
	if !bytes.Contains(res.Stdout, []byte("Initialized empty Git repository")) {
		require.NoError(t, git.NewUnknownRunErr(res, errors.New("unexpected output")))
	}
	return Open(t, path)
}

func InitBare(t *testing.T, path string) {
	ctx := context.Background()
	res, err := exec.Git{
		Path: path,
	}.Run(ctx, "init", "--bare")
	if err != nil {
		require.NoError(t, git.NewUnknownRunErr(res, err))
	}
	if !bytes.Contains(res.Stdout, []byte("Initialized empty Git repository")) {
		require.NoError(t, git.NewUnknownRunErr(res, errors.New("unexpected output")))
	}
}
